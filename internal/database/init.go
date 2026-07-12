// Package database
package database

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func NewDB(DBURL string) (*Queries, error) {
	slog.Info("initializing PG database", "URL", DBURL)

	pool, err := pgxpool.New(context.Background(), DBURL)
	if err != nil {
		slog.Error("database initialization failed", "URL", DBURL, "err", err)
		return nil, err
	}

	// we create a context so the ping call doesnt hang indefinitely in case the pg server is not reachable
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		slog.Error("failed to ping database", "err", err)
		return nil, err
	}

	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	err = runMigration(db, goose.DialectPostgres)
	if err != nil {
		pool.Close()
		slog.Error("migration error", "err", err)
		return nil, err
	}

	return New(pool), nil
}
