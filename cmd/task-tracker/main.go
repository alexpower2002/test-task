package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	tasktracker "mkk-luna-test-task/internal/app/task-tracker"
)

type RRR interface {
	Register() error
	Resolve(ctx context.Context) error
	Release() error
}

func handleSignals(app RRR, cancel context.CancelFunc) {
	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	for range done {
		log.Println("shutting server down...")

		cancel()

		err := app.Release()

		if err != nil {
			log.Fatalf("an error occurred during server shutdown: %v\n", err)
		}

		log.Println("server exited successfully")

		os.Exit(0)
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	var app RRR

	app = tasktracker.NewApp()

	go handleSignals(app, cancel)

	if err := app.Register(); err != nil {
		log.Println(err)

		os.Exit(1)
	}

	log.Println("initialized successfully")

	if err := app.Resolve(ctx); err != nil {
		log.Fatalln(err)
	}
}
