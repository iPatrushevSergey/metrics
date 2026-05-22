package main

import (
	"flag"
	"log"

	"github.com/iPatrushevSergey/metrics/app/internal/pkg/migrate"
)

func main() {
	dsn := flag.String("d", "", "database dsn")
	dir := flag.String("dir", migrate.MigrationsMetricsDir(), "path to migration files")
	flag.Parse()

	if *dsn == "" {
		log.Fatal("migrate: -d is required")
	}

	if err := migrate.Up(*dsn, *dir); err != nil {
		log.Fatalf("migrate: %v", err)
	}
}
