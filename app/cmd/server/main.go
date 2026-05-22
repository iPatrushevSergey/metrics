package main

import (
	"log"

	"github.com/iPatrushevSergey/metrics/app/cmd/server/bootstrap"
)

func main() {
	printBuildInfo()

	if err := bootstrap.Run(); err != nil {
		log.Fatalf("server: %v", err)
	}
}
