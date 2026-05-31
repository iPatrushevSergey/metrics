// Package router registers metrics HTTP routes on a Gin engine.
package router

import (
	"github.com/gin-gonic/gin"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/port"
	presport "github.com/iPatrushevSergey/metrics/app/internal/server/metrics/presentation/port"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/presentation/http/handler"
)

// RegisterRoutes registers metrics endpoints on the router.
func RegisterRoutes(r *gin.Engine, ucFactory presport.UseCaseFactory, log port.Logger) {
	metricHandler := handler.NewMetricHandler(ucFactory, log)
	r.GET("/ping", metricHandler.PingDB)
	r.GET("/", metricHandler.GetAll)
	r.POST("/update/", metricHandler.UpdateJSON)
	r.POST("/updates/", metricHandler.UpdatesJSON)
	r.POST("/value/", metricHandler.GetJSON)
	r.POST("/update/:type/:name/:value/", metricHandler.Update)
	r.GET("/value/:type/:name/", metricHandler.GetValue)
}
