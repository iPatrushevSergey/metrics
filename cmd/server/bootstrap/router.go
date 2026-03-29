package bootstrap

import (
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	gojson "github.com/goccy/go-json"
	"github.com/iPatrushevSergey/metrics/internal/config"
	"github.com/iPatrushevSergey/metrics/internal/handler"
	"github.com/iPatrushevSergey/metrics/internal/logger"
	"github.com/iPatrushevSergey/metrics/internal/middleware"
)

// GinJSONSerializer implements JSON marshal/unmarshal for the router.
type GinJSONSerializer struct{}

// Serialize marshals data to JSON bytes.
func (g *GinJSONSerializer) Serialize(c *gin.Context, data any) ([]byte, error) {
	return gojson.Marshal(data)
}

// Deserialize unmarshals JSON bytes into v.
func (g *GinJSONSerializer) Deserialize(c *gin.Context, data []byte, v any) error {
	return gojson.Unmarshal(data, v)
}

// SetupRouter configures and returns the HTTP router with all routes and middleware
func SetupRouter(metricHandler *handler.MetricHandler, cfg config.ServerConfig, log logger.Logger) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.GzipGinMiddleware(log))
	router.Use(middleware.HashMiddleware(cfg.Key))
	router.Use(middleware.LoggerMiddleware())
	router.Use(func(c *gin.Context) {
		c.Set("json.Serializer", &GinJSONSerializer{})
		c.Next()
	})

	pprof.Register(router)

	router.GET("/ping", metricHandler.PingDB)
	router.GET("/", metricHandler.GetAll)
	router.POST("/update/", metricHandler.UpdateJSON)
	router.POST("/updates/", metricHandler.UpdatesJSON)
	router.POST("/value/", metricHandler.GetJSON)
	router.POST("/update/:type/:name/:value/", metricHandler.Update)
	router.GET("/value/:type/:name/", metricHandler.GetValue)

	return router
}
