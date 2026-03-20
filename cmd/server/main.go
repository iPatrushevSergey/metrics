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
		log.Printf("error load config: %v\n%v", cfg, err)
		exit(1)
	}

	// Initialize and run the application
	if err := bootstrap.Run(cfg); err != nil {
		log.Printf("error running application: %v", err)
		exit(1)
	}
}
