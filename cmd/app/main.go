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

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"test-task/internal/config"
	grpcsvc "test-task/internal/grpc"
	"test-task/internal/rates"
	"test-task/internal/storage"
	"test-task/internal/utils"
	"test-task/proto/ratespb"
)

type RateStore interface {
	grpcsvc.RateRepository
	Close() error
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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	store, err := connectWithRetry(ctx, cfg)

	if err != nil {
		return err
	}

	defer func() {
		_ = store.Close()
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
	rateServer := grpcsvc.NewRateServer(store, provider)

	ratespb.RegisterRateServiceServer(grpcServer, rateServer)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":"+cfg.GRPCPort)

	if err != nil {
		return fmt.Errorf("listen on gRPC port %s: %w", cfg.GRPCPort, err)
	}

	serveErr := make(chan error, 1)

	go func() {
		serveErr <- grpcServer.Serve(listener)
	}()

	select {
	case <-ctx.Done():
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

func connectWithRetry(ctx context.Context, cfg config.Config) (RateStore, error) {
	delay := cfg.DBConnectDelay

	var lastErr error

	for attempt := 1; attempt <= cfg.DBConnectAttempts; attempt++ {
		store, err := storage.NewPostgres(ctx, cfg.PostgresDSN)

		if err == nil {
			return store, nil
		}

		lastErr = err

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}

	return nil, fmt.Errorf("connect to database: %w", lastErr)
}
