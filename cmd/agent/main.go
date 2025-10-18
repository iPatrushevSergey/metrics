package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/iPatrushevSergey/metrics/internal/agent"
)

func main() {
	cfg := agent.Config{
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
		ServerAddress:  "http://127.0.0.1:8080",
	}

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
