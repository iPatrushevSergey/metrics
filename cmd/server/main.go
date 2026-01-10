package main

import (
	"log"

	"github.com/iPatrushevSergey/metrics/cmd/server/bootstrap"
	"github.com/iPatrushevSergey/metrics/internal/config"
)

func main() {
	// Loading the config
	cfg, err := config.LoadServerConfig()
	if err != nil {
		log.Fatalf("error load config: %v\n%v", cfg, err)
	}

	// Initialize and run the application
	if err := bootstrap.Run(cfg); err != nil {
		log.Fatalf("error running application: %v", err)
	}
}
