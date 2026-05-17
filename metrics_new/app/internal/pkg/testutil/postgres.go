//go:build integration

package testutil

import (
	"context"
	"database/sql"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/jackc/pgx/v5/stdlib"
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

	applyMigrations(tb, dsn, migrationsDir())

	return pool
}

func applyMigrations(tb testing.TB, dsn, dir string) {
	tb.Helper()

	db, err := sql.Open("pgx", dsn)
	require.NoError(tb, err, "failed to open db for migrations")
	defer db.Close()

	require.NoError(tb, goose.SetDialect("postgres"), "failed to set migration dialect")
	require.NoError(tb, goose.Up(db, dir), "failed to apply migrations")
}

func migrationsDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "..", "..", "migrations", "metrics")
}
