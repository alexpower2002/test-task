package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"rates-service/internal/config"
	grpcsvc "rates-service/internal/grpc"
	"rates-service/internal/observability"
	"rates-service/internal/rates"
	"rates-service/internal/storage"
	"rates-service/internal/utils"
	"rates-service/proto/ratespb"
)

type AppLogger interface {
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Sync() error
}

type RateStore interface {
	grpcsvc.RateRepository
	Close() error
}

type AppMigrator interface {
	Up(ctx context.Context) error
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "app startup failed: %v\n", err)

		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load(os.Args[1:])

	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger, err := newLogger(cfg.LogLevel)

	if err != nil {
		return fmt.Errorf("build logger: %w", err)
	}

	defer func() {
		_ = logger.Sync()
	}()

	traceShutdown, err := observability.SetupTracing(logger, cfg.TraceExporter, cfg.ServiceName)

	if err != nil {
		return fmt.Errorf("setup tracing: %w", err)
	}

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		_ = traceShutdown.Shutdown(ctx)
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	store, err := connectWithRetry(ctx, cfg, logger)

	if err != nil {
		return err
	}

	defer func() {
		_ = store.Close()
	}()

	var migrator AppMigrator = storage.NewPostgresMigrator(cfg.PostgresDSN)

	if err := migrator.Up(ctx); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	var metrics grpcsvc.RequestMetrics = observability.NewPrometheusMetrics()

	metricsServer := observability.StartMetricsServer(":"+cfg.MetricsPort, logger)

	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		_ = metricsServer.Shutdown(shutdownCtx)
	}()

	var httpClient rates.HTTPClient = utils.NewRestyClient(cfg.HTTPTimeout)
	var fetcher rates.ProviderFetcher = rates.NewFetcher(httpClient, cfg.GrinexURL)

	askCalculator, err := newCalculator(cfg.Ask)

	if err != nil {
		return fmt.Errorf("build ask calculator: %w", err)
	}

	bidCalculator, err := newCalculator(cfg.Bid)

	if err != nil {
		return fmt.Errorf("build bid calculator: %w", err)
	}

	var provider grpcsvc.RateProvider = rates.NewProvider(fetcher, askCalculator, bidCalculator)

	grpcServer := grpc.NewServer()
	rateServer := grpcsvc.NewRateServer(store, provider, logger, metrics)

	ratespb.RegisterRateServiceServer(grpcServer, rateServer)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":"+cfg.GRPCPort)

	if err != nil {
		return fmt.Errorf("listen on gRPC port %s: %w", cfg.GRPCPort, err)
	}

	serveErr := make(chan error, 1)

	go func() {
		logger.Info("gRPC server started", zap.String("port", cfg.GRPCPort))

		serveErr <- grpcServer.Serve(listener)
	}()

	logger.Info("metrics server started", zap.String("port", cfg.MetricsPort))

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-serveErr:
		if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			return fmt.Errorf("serve gRPC: %w", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	done := make(chan struct{})

	go func() {
		defer close(done)

		grpcServer.GracefulStop()
	}()

	select {
	case <-done:
	case <-shutdownCtx.Done():
		grpcServer.Stop()
	}

	_ = metricsServer.Shutdown(shutdownCtx)

	return nil
}

func newCalculator(calculation config.Calculation) (rates.Calculator, error) {
	switch calculation.Method {
	case config.MethodTopN:
		return rates.NewTopCalculator(calculation.N), nil
	case config.MethodAvgNM:
		return rates.NewAverageCalculator(calculation.N, calculation.M), nil
	default:
		return nil, fmt.Errorf("unsupported calculation method %q", calculation.Method)
	}
}

func connectWithRetry(ctx context.Context, cfg config.Config, logger AppLogger) (RateStore, error) {
	delay := cfg.DBConnectDelay

	var lastErr error

	for attempt := 1; attempt <= cfg.DBConnectAttempts; attempt++ {
		store, err := storage.NewPostgres(ctx, cfg.PostgresDSN)

		if err == nil {
			logger.Info("database connection established", zap.Int("attempt", attempt))

			return store, nil
		}

		lastErr = err

		logger.Warn("database connection attempt failed",
			zap.Int("attempt", attempt),
			zap.Int("max_attempts", cfg.DBConnectAttempts),
			zap.Error(err),
		)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}

	return nil, fmt.Errorf("connect to database: %w", lastErr)
}

func newLogger(level string) (AppLogger, error) {
	var zapLevel zap.AtomicLevel

	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		return nil, err
	}

	loggerConfig := zap.NewProductionConfig()

	loggerConfig.Level = zapLevel

	return loggerConfig.Build()
}
