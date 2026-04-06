package observability

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

type Logger interface {
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
}

type Shutdowner interface {
	Shutdown(ctx context.Context) error
}

type PrometheusMetrics struct {
	RequestsTotal  *prometheus.CounterVec
	RequestLatency *prometheus.HistogramVec
}

type metricsServer struct {
	server *http.Server
	once   sync.Once
}

type shutdownFunc func(ctx context.Context) error

func NewPrometheusMetrics() *PrometheusMetrics {
	return &PrometheusMetrics{
		RequestsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "rates_requests_total",
			Help: "Total number of gRPC requests processed by the service.",
		}, []string{"method", "status"}),
		RequestLatency: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "rates_request_duration_seconds",
			Help:    "Latency of gRPC requests.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method"}),
	}
}

func (m *PrometheusMetrics) ObserveRequest(method, status string, duration time.Duration) {
	m.RequestsTotal.WithLabelValues(method, status).Inc()
	m.RequestLatency.WithLabelValues(method).Observe(duration.Seconds())
}

func StartMetricsServer(addr string, logger Logger) Shutdowner {
	server := &http.Server{
		Addr:              addr,
		Handler:           promhttp.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("metrics server failed", zap.Error(err))
		}
	}()

	return &metricsServer{server: server}
}

func (s *metricsServer) Shutdown(ctx context.Context) error {
	var shutdownErr error

	s.once.Do(func() {
		shutdownErr = s.server.Shutdown(ctx)
	})

	return shutdownErr
}

func SetupTracing(logger Logger, exporter, serviceName string) (Shutdowner, error) {
	if exporter == "none" {
		return shutdownFunc(func(context.Context) error { return nil }), nil
	}

	traceExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			"",
			attribute.String("service.name", serviceName),
		)),
	)

	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	logger.Info("OpenTelemetry tracing initialized", zap.String("exporter", exporter))

	return shutdownFunc(provider.Shutdown), nil
}

func (f shutdownFunc) Shutdown(ctx context.Context) error {
	return f(ctx)
}
