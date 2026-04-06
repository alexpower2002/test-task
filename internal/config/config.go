package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	MethodTopN  = "topN"
	MethodAvgNM = "avgNM"
)

type Calculation struct {
	Method string
	N      int
	M      int
}

type Config struct {
	GRPCPort          string
	MetricsPort       string
	PostgresDSN       string
	GrinexURL         string
	HTTPTimeout       time.Duration
	ShutdownTimeout   time.Duration
	DBConnectAttempts int
	DBConnectDelay    time.Duration
	LogLevel          string
	ServiceName       string
	TraceExporter     string
	Ask               Calculation
	Bid               Calculation
}

func Load(args []string) (Config, error) {
	fs := flag.NewFlagSet("rates-service", flag.ContinueOnError)

	cfg := Config{}

	fs.StringVar(&cfg.GRPCPort, "grpc-port", envString("APP_GRPC_PORT", "9001"), "gRPC listen port")
	fs.StringVar(&cfg.MetricsPort, "metrics-port", envString("APP_METRICS_PORT", "9090"), "Prometheus metrics port")
	fs.StringVar(&cfg.PostgresDSN, "postgres-dsn", envString("APP_POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/rates?sslmode=disable"), "PostgreSQL connection DSN")
	fs.StringVar(&cfg.GrinexURL, "grinex-url", envString("APP_GRINEX_URL", "https://grinex.io/api/v1/spot/depth?symbol=usdta7a5"), "Grinex depth API URL")
	fs.StringVar(&cfg.LogLevel, "log-level", envString("APP_LOG_LEVEL", "info"), "zap log level")
	fs.StringVar(&cfg.ServiceName, "service-name", envString("APP_SERVICE_NAME", "rates-service"), "OpenTelemetry service name")
	fs.StringVar(&cfg.TraceExporter, "trace-exporter", envString("APP_TRACE_EXPORTER", "stdout"), "OpenTelemetry exporter: stdout or none")

	httpTimeout := fs.String("http-timeout", envString("APP_HTTP_TIMEOUT", "5s"), "HTTP client timeout")
	shutdownTimeout := fs.String("shutdown-timeout", envString("APP_SHUTDOWN_TIMEOUT", "10s"), "Graceful shutdown timeout")
	dbConnectDelay := fs.String("db-connect-delay", envString("APP_DB_CONNECT_DELAY", "1s"), "Delay between database connection attempts")
	fs.IntVar(&cfg.DBConnectAttempts, "db-connect-attempts", envInt("APP_DB_CONNECT_ATTEMPTS", 10), "Database connection attempts")

	fs.StringVar(&cfg.Ask.Method, "ask-method", envString("APP_ASK_METHOD", MethodTopN), "Ask calculation method: topN or avgNM")
	fs.IntVar(&cfg.Ask.N, "ask-n", envInt("APP_ASK_N", 0), "Ask calculation start position")
	fs.IntVar(&cfg.Ask.M, "ask-m", envInt("APP_ASK_M", 0), "Ask calculation end position")

	fs.StringVar(&cfg.Bid.Method, "bid-method", envString("APP_BID_METHOD", MethodTopN), "Bid calculation method: topN or avgNM")
	fs.IntVar(&cfg.Bid.N, "bid-n", envInt("APP_BID_N", 0), "Bid calculation start position")
	fs.IntVar(&cfg.Bid.M, "bid-m", envInt("APP_BID_M", 0), "Bid calculation end position")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	var err error

	if cfg.HTTPTimeout, err = time.ParseDuration(*httpTimeout); err != nil {
		return Config{}, fmt.Errorf("parse http-timeout: %w", err)
	}

	if cfg.ShutdownTimeout, err = time.ParseDuration(*shutdownTimeout); err != nil {
		return Config{}, fmt.Errorf("parse shutdown-timeout: %w", err)
	}

	if cfg.DBConnectDelay, err = time.ParseDuration(*dbConnectDelay); err != nil {
		return Config{}, fmt.Errorf("parse db-connect-delay: %w", err)
	}

	if err := validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func validate(cfg Config) error {
	if cfg.PostgresDSN == "" {
		return fmt.Errorf("postgres-dsn must not be empty")
	}

	if err := validateCalculation("ask", cfg.Ask); err != nil {
		return err
	}

	if err := validateCalculation("bid", cfg.Bid); err != nil {
		return err
	}

	if cfg.DBConnectAttempts < 1 {
		return fmt.Errorf("db-connect-attempts must be >= 1")
	}

	if cfg.GRPCPort == "" || cfg.MetricsPort == "" {
		return fmt.Errorf("ports must not be empty")
	}

	if cfg.TraceExporter != "stdout" && cfg.TraceExporter != "none" {
		return fmt.Errorf("trace-exporter must be stdout or none")
	}

	return nil
}

func validateCalculation(side string, calculation Calculation) error {
	if calculation.N < 0 {
		return fmt.Errorf("%s N must be >= 0", side)
	}

	switch calculation.Method {
	case MethodTopN:
		return nil
	case MethodAvgNM:
		if calculation.M < calculation.N {
			return fmt.Errorf("%s M must be >= N for avgNM", side)
		}

		return nil
	default:
		return fmt.Errorf("%s method must be one of %s, %s", side, MethodTopN, MethodAvgNM)
	}
}

func envString(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}

	return fallback
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))

	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)

	if err != nil {
		return fallback
	}

	return parsed
}
