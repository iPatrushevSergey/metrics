package main

import (
	"log"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/cmd/agent/bootstrap"
)

func main() {
	printBuildInfo()

	if err := bootstrap.Run(); err != nil {
		log.Fatalf("agent: %v", err)
	}
}
