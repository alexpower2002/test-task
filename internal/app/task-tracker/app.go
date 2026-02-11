package tasktracker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"mkk-luna-test-task/internal/email"
	"mkk-luna-test-task/internal/repository"
	"mkk-luna-test-task/internal/user"
	"mkk-luna-test-task/internal/utils"

	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	commentcreatehandler "mkk-luna-test-task/internal/task/comment/create"
	commentlisthandler "mkk-luna-test-task/internal/task/comment/list"
	taskcreatehandler "mkk-luna-test-task/internal/task/create"
	taskedithandler "mkk-luna-test-task/internal/task/edit"
	taskhistorylisthandler "mkk-luna-test-task/internal/task/history/list"
	tasklisthandler "mkk-luna-test-task/internal/task/list"
	teamcreatehandler "mkk-luna-test-task/internal/team/create"
	teamlisthandler "mkk-luna-test-task/internal/team/list"
	teaminvitehandler "mkk-luna-test-task/internal/team/member/invite"
	loginhandler "mkk-luna-test-task/internal/user/login"
	registerhandler "mkk-luna-test-task/internal/user/register"
)

const (
	timeout = 5 * time.Second

	circuitBreakerDefaultMaxFailures = 5
	circuitBreakerDefaultTimeout     = 10 * time.Second
)

type app struct {
	server http.Server
}

func NewApp() *app {
	return &app{
		server: http.Server{},
	}
}

func (a *app) Register() error {
	envs, err := utils.LoadEnvs()

	if err != nil {
		return err
	}

	if envs.Port == 0 {
		return fmt.Errorf("empty port")
	}

	if err := repository.RunMigrations(envs.MySQL, envs.MigrationsPath); err != nil {
		return err
	}

	db, err := repository.NewMySqlClient(envs.MySQL)

	if err != nil {
		return err
	}

	repo := repository.NewMysql(db)

	redisClient, err := repository.NewRedisClient(envs.Redis)

	if err != nil {
		return err
	}

	redisTTL := envs.RedisTTL

	redisRepo := repository.NewRedis(redisClient, redisTTL)

	metrics, err := utils.InitMetrics()

	if err != nil {
		return err
	}

	metrics.IncStartup()

	chiRouter := chi.NewRouter()

	chiRouter.Handle("/metrics", promhttp.Handler())
	chiRouter.Handle("/health", utils.NewHealthcheckHandler())

	userGetter := user.NewJwtUserFromRequestGetter([]byte(envs.JwtSecret))

	rateLimitingMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			utils.NewRateLimitingMiddleware(next).Handle(w, r)
		})
	}

	userGetterMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user.NewUserGetterMiddleware(userGetter, next).Handle(w, r)
		})
	}

	metricsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			utils.NewMetricsMiddlewareHandler(metrics, next).Handle(w, r)
		})
	}

	registerExec := registerhandler.NewExecutor(repo)

	chiRouter.Post("/api/v1/register", func(w http.ResponseWriter, r *http.Request) {
		h := http.HandlerFunc(registerhandler.NewHandler(registerExec).Handle)

		utils.ChainMiddlewares(h, rateLimitingMiddleware, metricsMiddleware).ServeHTTP(w, r)
	})

	loginExec := loginhandler.NewExecutor(repo, []byte(envs.JwtSecret), envs.JwtExpiration)

	chiRouter.Post("/api/v1/login", func(w http.ResponseWriter, r *http.Request) {
		h := http.HandlerFunc(loginhandler.NewHandler(loginExec).Handle)

		utils.ChainMiddlewares(h, rateLimitingMiddleware, metricsMiddleware).ServeHTTP(w, r)
	})

	teamCreateExec := teamcreatehandler.NewExecutor(repo)

	chiRouter.Post("/api/v1/teams", func(w http.ResponseWriter, r *http.Request) {
		h := http.HandlerFunc(teamcreatehandler.NewHandler(teamCreateExec).Handle)

		utils.ChainMiddlewares(h, rateLimitingMiddleware, userGetterMiddleware, metricsMiddleware).ServeHTTP(w, r)
	})

	teamListExec := teamlisthandler.NewExecutor(repo)

	chiRouter.Get("/api/v1/teams", func(w http.ResponseWriter, r *http.Request) {
		h := http.HandlerFunc(teamlisthandler.NewHandler(teamListExec).Handle)

		utils.ChainMiddlewares(h, rateLimitingMiddleware, userGetterMiddleware, metricsMiddleware).ServeHTTP(w, r)
	})

	circuitBreakerClient := utils.NewCircuitBreakerHTTPClient(
		circuitBreakerDefaultMaxFailures,
		circuitBreakerDefaultTimeout,
	)

	mockEmailSender := email.NewStubSender(circuitBreakerClient)

	inviteExec := teaminvitehandler.NewExecutor(repo, repo, mockEmailSender)

	chiRouter.Post("/api/v1/teams/{id}/invite", func(w http.ResponseWriter, r *http.Request) {
		h := http.HandlerFunc(teaminvitehandler.NewHandler(inviteExec).Handle)

		utils.ChainMiddlewares(h, rateLimitingMiddleware, userGetterMiddleware, metricsMiddleware).ServeHTTP(w, r)
	})

	taskCreateExec := taskcreatehandler.NewExecutor(repo, repo)

	chiRouter.Post("/api/v1/tasks", func(w http.ResponseWriter, r *http.Request) {
		h := http.HandlerFunc(taskcreatehandler.NewHandler(taskCreateExec).Handle)

		utils.ChainMiddlewares(h, rateLimitingMiddleware, userGetterMiddleware, metricsMiddleware).ServeHTTP(w, r)
	})

	taskListExec := tasklisthandler.NewExecutor(repo, repo, redisRepo, redisRepo)

	chiRouter.Get("/api/v1/tasks", func(w http.ResponseWriter, r *http.Request) {
		h := http.HandlerFunc(tasklisthandler.NewHandler(taskListExec).Handle)

		utils.ChainMiddlewares(h, rateLimitingMiddleware, userGetterMiddleware, metricsMiddleware).ServeHTTP(w, r)
	})

	taskEditExec := taskedithandler.NewExecutor(repo, repo, repo, repo, redisRepo)

	chiRouter.Put("/api/v1/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		h := http.HandlerFunc(taskedithandler.NewHandler(taskEditExec).Handle)

		utils.ChainMiddlewares(h, rateLimitingMiddleware, userGetterMiddleware, metricsMiddleware).ServeHTTP(w, r)
	})

	taskHistoryListExec := taskhistorylisthandler.NewExecutor(repo, repo)

	chiRouter.Get("/api/v1/tasks/{id}/history", func(w http.ResponseWriter, r *http.Request) {
		h := http.HandlerFunc(taskhistorylisthandler.NewHandler(taskHistoryListExec).Handle)

		utils.ChainMiddlewares(h, rateLimitingMiddleware, userGetterMiddleware, metricsMiddleware).ServeHTTP(w, r)
	})

	commentCreateExec := commentcreatehandler.NewExecutor(repo, repo)

	chiRouter.Post("/api/v1/tasks/{id}/comments", func(w http.ResponseWriter, r *http.Request) {
		h := http.HandlerFunc(commentcreatehandler.NewHandler(commentCreateExec).Handle)

		utils.ChainMiddlewares(h, rateLimitingMiddleware, userGetterMiddleware, metricsMiddleware).ServeHTTP(w, r)
	})

	commentListExec := commentlisthandler.NewExecutor(repo)

	chiRouter.Get("/api/v1/tasks/{id}/comments", func(w http.ResponseWriter, r *http.Request) {
		h := http.HandlerFunc(commentlisthandler.NewHandler(commentListExec).Handle)

		utils.ChainMiddlewares(h, rateLimitingMiddleware, userGetterMiddleware, metricsMiddleware).ServeHTTP(w, r)
	})

	a.server.Handler = chiRouter
	a.server.Addr = ":" + strconv.Itoa(envs.Port)

	return nil
}

func (a *app) Resolve(context.Context) error {
	errChan := make(chan error)

	go func() {
		log.Printf("server started on address: %s", a.server.Addr)

		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("listening err: %v", err)

			errChan <- err
		}
	}()

	return <-errChan
}

func (a *app) Release() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	return nil
}
