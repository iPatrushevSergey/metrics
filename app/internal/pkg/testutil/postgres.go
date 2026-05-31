//go:build integration

package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/iPatrushevSergey/metrics/app/internal/pkg/migrate"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	testDBName = "metrics_test"
	testDBUser = "test"
	testDBPass = "test"
)

// SetupPostgres starts a PostgreSQL container, applies migrations and returns
// a ready-to-use pgxpool.Pool. The container is terminated when tb finishes.
func SetupPostgres(tb testing.TB) *pgxpool.Pool {
	tb.Helper()
	ctx := context.Background()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase(testDBName),
		tcpostgres.WithUsername(testDBUser),
		tcpostgres.WithPassword(testDBPass),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(tb, err, "failed to start postgres container")

	tb.Cleanup(func() {
		require.NoError(tb, pgContainer.Terminate(ctx), "failed to terminate postgres container")
	})

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(tb, err, "failed to get connection string")

	poolCfg, err := pgxpool.ParseConfig(dsn)
	require.NoError(tb, err)
	poolCfg.MaxConns = 5
	poolCfg.MinConns = 1

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	require.NoError(tb, err, "failed to create pool")

	tb.Cleanup(func() { pool.Close() })

	require.NoError(tb, migrate.Up(dsn, migrate.MigrationsMetricsDir()), "failed to apply migrations")

	return pool
}
