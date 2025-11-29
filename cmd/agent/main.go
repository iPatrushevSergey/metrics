package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/iPatrushevSergey/metrics/internal/agent"
	"github.com/iPatrushevSergey/metrics/internal/config"
)

func main() {
	cfg, err := config.LoadAgentConfig()
	if err != nil {
		log.Fatalf("error load config: %v", err)
	}

	log.Printf("Starting agent with config: %+v\n", cfg)
	log.Println("Sending metrics to the server", cfg.Address)

	a := agent.NewAgent(cfg)

	// The pattern "Graceful Shutdown"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // The pattern "belt and suspenders"

	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		a.PollMetrics(ctx)
	}()
	go func() {
		defer wg.Done()
		a.ReportMetrics(ctx)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("The completion signal has been received, starting the stop...")
	cancel()

	wg.Wait()
	log.Println("The agent has been stopped")
}
