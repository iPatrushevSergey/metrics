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
	"go.uber.org/zap"

	gojson "github.com/goccy/go-json"

	"github.com/iPatrushevSergey/metrics/internal/handler"
	"github.com/iPatrushevSergey/metrics/internal/repository/inmemory"
	"github.com/iPatrushevSergey/metrics/internal/service"

	"github.com/iPatrushevSergey/metrics/internal/config"
	"github.com/iPatrushevSergey/metrics/internal/logger"
)

type GinJSONSerializer struct{}

func (g *GinJSONSerializer) Serialize(c *gin.Context, data interface{}) ([]byte, error) {
	return gojson.Marshal(data)
}

func (g *GinJSONSerializer) Deserialize(c *gin.Context, data []byte, v interface{}) error {
	return gojson.Unmarshal(data, v)
}

func main() {
	cfg, err := config.LoadServerConfig()
	if err != nil {
		log.Fatalf("error load config: %v\n%v", cfg, err)
	}

	initializedLogger, err := logger.Initialize(cfg.LogLevel)
	if err != nil {
		log.Fatalf("error initialize logger: %v", err)
	}
	defer initializedLogger.Sync()

	logger.Log.Info("starting server with config", zap.Object("cfg details", &cfg))

	repo := inmemory.NewMemStorageMetricRepository()
	metricService := service.NewMetricService(repo)
	metricHandler := handler.NewMetricHandler(metricService)

	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Set("json.Serializer", &GinJSONSerializer{})
		c.Next()
	})
	router.Use(gin.Recovery())
	router.Use(logger.ZapLogger())

	router.GET("/", metricHandler.GetAll)
	router.POST("/update", metricHandler.UpdateJSON)
	router.POST("/value", metricHandler.GetJSON)
	router.POST("/update/:type/:name/:value", metricHandler.Update)
	router.GET("/value/:type/:name", metricHandler.GetValue)

	server := &http.Server{
		Addr:    cfg.Address,
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Error("Server failed to start", zap.Error(err))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Log.Info("The completion signal has been received, starting the stop...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger.Log.Info("Shutting down server...")
	if err := server.Shutdown(ctx); err != nil {
		logger.Log.Error("Server shutdown failed", zap.Error(err))
		os.Exit(1)
	}

	logger.Log.Info("Server stopped gracefully")
}
