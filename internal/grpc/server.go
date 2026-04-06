package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"test-task/internal/rates"
	"test-task/proto/ratespb"
)

type RateRepository interface {
	SaveRate(ctx context.Context, rate *rates.Rate) error
	Ping(ctx context.Context) error
}

type RateProvider interface {
	FetchRates(ctx context.Context) (*rates.Rate, error)
}

type RateServer struct {
	ratespb.UnimplementedRateServiceServer
	repository RateRepository
	provider   RateProvider
}

func NewRateServer(repository RateRepository, provider RateProvider) *RateServer {
	return &RateServer{
		repository: repository,
		provider:   provider,
	}
}

func (s *RateServer) GetRates(ctx context.Context, _ *ratespb.GetRatesRequest) (*ratespb.GetRatesResponse, error) {
	rate, err := s.provider.FetchRates(ctx)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "fetch rates: %v", err)
	}

	if err := s.repository.SaveRate(ctx, rate); err != nil {
		return nil, status.Errorf(codes.Internal, "save rate: %v", err)
	}

	return &ratespb.GetRatesResponse{
		Ask:       rate.Ask,
		Bid:       rate.Bid,
		Timestamp: rate.Timestamp.Format(time.RFC3339),
	}, nil
}

func (s *RateServer) HealthCheck(ctx context.Context, _ *ratespb.HealthCheckRequest) (*ratespb.HealthCheckResponse, error) {
	if err := s.repository.Ping(ctx); err != nil {
		return nil, status.Errorf(codes.Unavailable, "database unavailable: %v", err)
	}

	return &ratespb.HealthCheckResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil
}
