package main

import (
	"log"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/cmd/server/bootstrap"
)

func main() {
	if err := bootstrap.Run(); err != nil {
		log.Fatalf("server: %v", err)
	}
}
