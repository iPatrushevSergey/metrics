package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"go.uber.org/zap"

	"github.com/iPatrushevSergey/metrics/internal/agent"
	"github.com/iPatrushevSergey/metrics/internal/config"
	"github.com/iPatrushevSergey/metrics/internal/logger"
)

func main() {
	cfg, err := config.LoadAgentConfig()
	if err != nil {
		log.Fatalf("error load config: %v", err)
	}

	// Initialize zap logger
	initializedLogger, err := logger.Initialize(cfg.LogLevel)
	if err != nil {
		log.Fatalf("error initializing logger: %v", err)
	}
	defer initializedLogger.Sync()

	loggerAdapter := logger.NewZapLoggerAdapter(initializedLogger)

	loggerAdapter.Info("Starting agent", zap.Object("config", &cfg))
	loggerAdapter.Info("Sending metrics to the server", zap.String("address", cfg.Address))

	a, err := agent.NewAgent(cfg, loggerAdapter)
	if err != nil {
		loggerAdapter.Fatal("Failed to create agent", zap.Error(err))
	}

	// The pattern "Graceful Shutdown"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // The pattern "belt and suspenders"

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		a.PollMetrics(ctx)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.PollGopsutilMetrics(ctx)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.ReportMetrics(ctx)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	loggerAdapter.Info("The completion signal has been received, starting the stop...")
	cancel()

	wg.Wait()
	a.Stop()
	loggerAdapter.Info("The agent has been stopped")
}
