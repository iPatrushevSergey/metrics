package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PoolConfig defines PostgreSQL pool settings.
type PoolConfig struct {
	URI         string        `mapstructure:"uri"`
	MaxConns    int32         `mapstructure:"max_conns"`
	MinConns    int32         `mapstructure:"min_conns"`
	MaxConnLife time.Duration `mapstructure:"max_conn_life"`
	MaxConnIdle time.Duration `mapstructure:"max_conn_idle"`
	HealthCheck time.Duration `mapstructure:"health_check"`
}

// RetryConfig defines transactor retry settings for retriable DB errors.
type RetryConfig struct {
	MaxRetries int           `mapstructure:"max_retries"`
	BaseDelay  time.Duration `mapstructure:"base_delay"`
	MaxDelay   time.Duration `mapstructure:"max_delay"`
}

// Config groups adapter-level postgres settings.
type Config struct {
	Pool  PoolConfig  `mapstructure:",squash"`
	Retry RetryConfig `mapstructure:"retry"`
}

// NewPool creates a pgxpool.Pool from postgres adapter config.
func NewPool(ctx context.Context, cfg PoolConfig) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.URI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URI: %w", err)
	}

	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns
	poolCfg.MaxConnLifetime = cfg.MaxConnLife
	poolCfg.MaxConnIdleTime = cfg.MaxConnIdle
	poolCfg.HealthCheckPeriod = cfg.HealthCheck

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}
