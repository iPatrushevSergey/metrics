package main

import (
	"context"
	"database/sql"
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
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/iPatrushevSergey/metrics/internal/config"
	"github.com/iPatrushevSergey/metrics/internal/filestorage"
	"github.com/iPatrushevSergey/metrics/internal/handler"
	"github.com/iPatrushevSergey/metrics/internal/logger"
	"github.com/iPatrushevSergey/metrics/internal/middleware"
	"github.com/iPatrushevSergey/metrics/internal/repository"
	"github.com/iPatrushevSergey/metrics/internal/repository/inmemory"
	"github.com/iPatrushevSergey/metrics/internal/repository/postgres"
	"github.com/iPatrushevSergey/metrics/internal/service"
)

const (
	// dbPingTimeout timeout for checking the connection to the database
	dbPingTimeout = 3 * time.Second
	// serverShutdownTimeout timeout for graceful shutdown server
	serverShutdownTimeout = 5 * time.Second
)

type GinJSONSerializer struct{}

func (g *GinJSONSerializer) Serialize(c *gin.Context, data interface{}) ([]byte, error) {
	return gojson.Marshal(data)
}

func (g *GinJSONSerializer) Deserialize(c *gin.Context, data []byte, v interface{}) error {
	return gojson.Unmarshal(data, v)
}

func main() {
	// Loading the config
	cfg, err := config.LoadServerConfig()
	if err != nil {
		log.Fatalf("error load config: %v\n%v", cfg, err)
	}

	// Loading the logger
	initializedLogger, err := logger.Initialize(cfg.LogLevel)
	if err != nil {
		log.Fatalf("error initialize logger: %v", err)
	}
	defer initializedLogger.Sync()

	logger.Log.Debug("starting server with config", zap.Object("cfg details", &cfg))

	// Creating a repository
	var repo repository.MetricRepository
	var fs *filestorage.FileStorage
	var periodicSaver *filestorage.PeriodicSaver

	if cfg.DatabaseDSN != "" {
		// PostgreSQL
		logger.Log.Debug("Using PostgreSQL storage")

		if err := postgres.RunMigrations(cfg.DatabaseDSN); err != nil {
			logger.Log.Fatal("Failed to run migrations", zap.Error(err))
		}

		db, err := sql.Open("pgx", cfg.DatabaseDSN)
		if err != nil {
			logger.Log.Fatal("Failed to open DB connection", zap.Error(err))
		}
		defer db.Close()

		ctx, cancel := context.WithTimeout(context.Background(), dbPingTimeout)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			logger.Log.Fatal("Unable to connect to database", zap.Error(err))
		}

		repo = postgres.NewPostgresMetricRepository(db)
		fs = nil
	} else {
		// Inmemory
		logger.Log.Debug("Using In-Memory storage with FileStorage")

		memRepo := inmemory.NewMemStorageMetricRepository()
		fs = filestorage.NewFileStorage(cfg.FileStoragePath)

		if cfg.Restore {
			logger.Log.Debug("Restoring metrics from file", zap.String("path", cfg.FileStoragePath))
			restoredMetrics, err := fs.Load()
			if err != nil {
				logger.Log.Error("Failed to restore metrics", zap.Error(err))
			} else {
				ctxRestore := context.Background()
				for _, m := range restoredMetrics {
					if err := memRepo.Create(ctxRestore, m); err != nil {
						logger.Log.Warn("Failed to restore metric", zap.String("id", m.ID), zap.Error(err))
					}
				}
				logger.Log.Debug("Metrics restored successfully", zap.Int("count", len(restoredMetrics)))
			}
		}

		// Defining the inmemory type (sync or async)
		if cfg.StoreInterval == 0 {
			logger.Log.Debug("Sync saving mode enable (StorageInterval=0)")
			repo = repository.NewSyncFileRepository(memRepo, fs)
		} else {
			repo = memRepo
		}

		if cfg.StoreInterval > 0 {
			periodicSaver = filestorage.NewPeriodicSaver(repo, fs, cfg.StoreInterval)
			periodicSaver.Start()
		}
	}

	metricService := service.NewMetricService(repo)
	metricHandler := handler.NewMetricHandler(metricService, logger.NewZapLoggerAdapter(initializedLogger))

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.GzipGinMiddleware())
	router.Use(middleware.LoggerMiddleware())
	router.Use(func(c *gin.Context) {
		c.Set("json.Serializer", &GinJSONSerializer{})
		c.Next()
	})

	router.GET("/ping", metricHandler.PingDB)
	router.GET("/", metricHandler.GetAll)
	router.POST("/update", metricHandler.UpdateJSON)
	router.POST("/updates", metricHandler.UpdatesJSON)
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
	logger.Log.Debug("The completion signal has been received, starting the stop...")

	ctx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeout)
	defer cancel()

	if cfg.DatabaseDSN == "" && fs != nil {
		if periodicSaver != nil {
			periodicSaver.Stop()
		}
		if err := filestorage.SaveOnShutdown(ctx, repo, fs); err != nil {
			logger.Log.Error("Failed to save metrics on shutdown", zap.Error(err))
		}
	}

	logger.Log.Debug("Shutting down server...")
	if err := server.Shutdown(ctx); err != nil {
		logger.Log.Error("Server shutdown failed", zap.Error(err))
		os.Exit(1)
	}

	logger.Log.Debug("Server stopped gracefully")
}
