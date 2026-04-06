package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
)

type PostgresMigrator struct {
	dsn string
}

func NewPostgresMigrator(dsn string) *PostgresMigrator {
	return &PostgresMigrator{dsn: dsn}
}

func (m *PostgresMigrator) Up(ctx context.Context) error {
	db, err := sql.Open("postgres", m.dsn)

	if err != nil {
		return fmt.Errorf("open postgres for migrations: %w", err)
	}

	defer func() {
		_ = db.Close()
	}()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping postgres before migrations: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})

	if err != nil {
		return fmt.Errorf("create postgres migration driver: %w", err)
	}

	sourceDriver, err := iofs.New(os.DirFS("migrations"), ".")

	if err != nil {
		return fmt.Errorf("create migration source: %w", err)
	}

	migrator, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)

	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	defer func() {
		_, _ = migrator.Close()
	}()

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}
