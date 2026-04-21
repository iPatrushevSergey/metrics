package main

import (
	"flag"
	"log"
)

func main() {
	dsn := flag.String("d", "", "database dsn")
	dir := flag.String("dir", "../migrations/metrics", "path to migration files")
	flag.Parse()

	if *dsn == "" {
		log.Fatal("migrate: -d is required")
	}

	if err := runMigrations(*dsn, *dir); err != nil {
		log.Fatalf("migrate: %v", err)
	}
}
