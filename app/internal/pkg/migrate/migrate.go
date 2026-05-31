// Package migrate applies SQL schema migrations for the metrics database.
package migrate

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/pressly/goose/v3"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// MigrationsMetricsDir returns the path to the migrations directory.
func MigrationsMetricsDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "..", "..", "migrations", "metrics")
}

// Up applies all pending migrations to the database.
func Up(dsn, dir string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	if err := goose.Up(db, dir); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	return nil
}
