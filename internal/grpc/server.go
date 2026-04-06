package grpc

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"rates-service/internal/rates"
	"rates-service/proto/ratespb"
)

type RateRepository interface {
	SaveRate(ctx context.Context, rate *rates.Rate) error
	Ping(ctx context.Context) error
}

type RateProvider interface {
	FetchRates(ctx context.Context) (*rates.Rate, error)
}

type Logger interface {
	Error(msg string, fields ...zap.Field)
}

type RequestMetrics interface {
	ObserveRequest(method, status string, duration time.Duration)
}

type RateServer struct {
	ratespb.UnimplementedRateServiceServer
	repository RateRepository
	provider   RateProvider
	logger     Logger
	metrics    RequestMetrics
}

func NewRateServer(repository RateRepository, provider RateProvider, logger Logger, metrics RequestMetrics) *RateServer {
	return &RateServer{
		repository: repository,
		provider:   provider,
		logger:     logger,
		metrics:    metrics,
	}
}

func (s *RateServer) GetRates(ctx context.Context, _ *ratespb.GetRatesRequest) (*ratespb.GetRatesResponse, error) {
	startedAt := time.Now()

	ctx, span := otel.Tracer("rates-service/grpc").Start(ctx, "grpc.GetRates")
	defer span.End()

	rate, err := s.provider.FetchRates(ctx)

	if err != nil {
		s.observe("GetRates", "error", startedAt)
		s.logger.Error("fetch rates failed", zap.Error(err))

		return nil, status.Errorf(codes.Internal, "fetch rates: %v", err)
	}

	if err := s.repository.SaveRate(ctx, rate); err != nil {
		s.observe("GetRates", "error", startedAt)
		s.logger.Error("save rate failed", zap.Error(err))

		return nil, status.Errorf(codes.Internal, "save rate: %v", err)
	}

	s.observe("GetRates", "ok", startedAt)

	return &ratespb.GetRatesResponse{
		Ask:       rate.Ask,
		Bid:       rate.Bid,
		Timestamp: rate.Timestamp.Format(time.RFC3339),
	}, nil
}

func (s *RateServer) HealthCheck(ctx context.Context, _ *ratespb.HealthCheckRequest) (*ratespb.HealthCheckResponse, error) {
	startedAt := time.Now()

	ctx, span := otel.Tracer("rates-service/grpc").Start(ctx, "grpc.HealthCheck")
	defer span.End()

	if err := s.repository.Ping(ctx); err != nil {
		s.observe("HealthCheck", "error", startedAt)
		s.logger.Error("health check failed", zap.Error(err))

		return nil, status.Errorf(codes.Unavailable, "database unavailable: %v", err)
	}

	s.observe("HealthCheck", "ok", startedAt)

	return &ratespb.HealthCheckResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (s *RateServer) observe(method, statusLabel string, startedAt time.Time) {
	s.metrics.ObserveRequest(method, statusLabel, time.Since(startedAt))
}
