package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func Open(ctx context.Context, databaseURL string, timeout time.Duration) (*Store, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	config.MaxConns = 12
	config.MinConns = 1
	config.MaxConnLifetime = 30 * time.Minute
	config.HealthCheckPeriod = 30 * time.Second
	openCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	pool, err := pgxpool.NewWithConfig(openCtx, config)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	if err := pool.Ping(openCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	if s != nil && s.pool != nil {
		s.pool.Close()
	}
}

func (s *Store) Ready(ctx context.Context) error {
	return s.pool.Ping(ctx)
}
