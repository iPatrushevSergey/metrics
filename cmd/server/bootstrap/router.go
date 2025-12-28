package bootstrap

import (
	"github.com/gin-gonic/gin"

	gojson "github.com/goccy/go-json"
	"github.com/iPatrushevSergey/metrics/internal/config"
	"github.com/iPatrushevSergey/metrics/internal/handler"
	"github.com/iPatrushevSergey/metrics/internal/middleware"
)

type GinJSONSerializer struct{}

func (g *GinJSONSerializer) Serialize(c *gin.Context, data interface{}) ([]byte, error) {
	return gojson.Marshal(data)
}

func (g *GinJSONSerializer) Deserialize(c *gin.Context, data []byte, v interface{}) error {
	return gojson.Unmarshal(data, v)
}

// SetupRouter configures and returns the HTTP router with all routes and middleware
func SetupRouter(metricHandler *handler.MetricHandler, cfg config.ServerConfig) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.GzipGinMiddleware())
	router.Use(middleware.HashMiddleware(cfg.Key))
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

	return router
}
