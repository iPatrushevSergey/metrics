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
	flags, err := config.AgentParseFlags()
	if err != nil {
		log.Fatalf("error parsing flags: %v", err)
	}
	log.Printf("Starting agent with options: %+v\n", flags)

	cfg := agent.Config{
		PollInterval:   flags.PollInterval,
		ReportInterval: flags.ReportInterval,
		ServerAddress:  "http://" + flags.NetAddress.String(),
	}
	log.Println("Running server on", cfg.ServerAddress)

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
