package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/iPatrushevSergey/metrics/internal/handler"
	"github.com/iPatrushevSergey/metrics/internal/repository/inmemory"
	"github.com/iPatrushevSergey/metrics/internal/service"

	"github.com/iPatrushevSergey/metrics/internal/config"
)

func main() {
	cfg, err := config.LoadServerConfig()
	if err != nil {
		log.Fatalf("error load config: %v", err)
		return
	}

	log.Printf("Starting server with config: %+v\n", cfg)

	repo := inmemory.NewMemStorageMetricRepository()
	metricService := service.NewMetricService(repo)
	metricHandler := handler.NewMetricHandler(metricService)

	router := gin.Default()

	router.GET("/", metricHandler.GetAll)
	router.POST("/update/:type/:name/:value", metricHandler.Update)
	router.GET("/value/:type/:name", metricHandler.Get)

	server := &http.Server{
		Addr:    cfg.Address,
		Handler: router,
	}

	go func() {
		log.Printf("Starting server, address: %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Server failed to start, error: %v", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("The completion signal has been received, starting the stop...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("Shutting down server...")
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown failed, error: %v", err)
		os.Exit(1)
	}

	log.Println("Server stopped gracefully")
}
