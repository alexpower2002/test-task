package storage

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"test-task/internal/rates"
)

type Postgres struct {
	db *sql.DB
}

func NewPostgres(ctx context.Context, dsn string) (*Postgres, error) {
	db, err := sql.Open("postgres", dsn)

	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()

		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &Postgres{db: db}, nil
}

func (p *Postgres) SaveRate(ctx context.Context, rate *rates.Rate) error {
	if _, err := p.db.ExecContext(
		ctx,
		`INSERT INTO rates (ask, bid, timestamp) VALUES ($1, $2, $3)`,
		rate.Ask,
		rate.Bid,
		rate.Timestamp,
	); err != nil {
		return fmt.Errorf("insert rate: %w", err)
	}

	return nil
}

func (p *Postgres) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

func (p *Postgres) Close() error {
	return p.db.Close()
}
