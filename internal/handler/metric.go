package handler

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"html/template"

	"github.com/gin-gonic/gin"
	"github.com/iPatrushevSergey/metrics/internal/logger"
	"github.com/iPatrushevSergey/metrics/internal/service"
	"go.uber.org/zap"
)

const metricsHTMLTemplate = `
	<html>
	<head><title>Metrics</title></head>
	<body>
		<h1>All metrics</h1>
		<ul>
			{{range .Metrics}}
				<li><b>{{.Name}}:</b> {{.Value}}</li>
			{{end}}
		</ul>
	</body>
	</html>
`

var metricsTemplate = template.Must(template.New("metrics").Parse(metricsHTMLTemplate))

type MetricHandler struct {
	metricService *service.MetricsService
}

func NewMetricHandler(s *service.MetricsService) *MetricHandler {
	return &MetricHandler{metricService: s}
}

func (h *MetricHandler) GetValue(c *gin.Context) {
	metricType := c.Param("type")
	metricName := c.Param("name")

	metric, err := h.metricService.GetValue(metricType, metricName)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.String(http.StatusNotFound, err.Error())
		} else if errors.Is(err, service.ErrBadMetricType) {
			c.String(http.StatusBadRequest, err.Error())
		} else {
			logger.Log.Error("Internal server error in GetValue", zap.Error(err))
			c.String(http.StatusInternalServerError, service.ErrInternal.Error())
		}
		return
	}

	c.String(http.StatusOK, "%s", metric)
}

func (h *MetricHandler) GetJSON(c *gin.Context) {
	var dto MetricDTO

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	if err := dto.UnmarshalJSON(body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	metricModel, err := h.metricService.GetMetric(dto.MType, dto.ID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if errors.Is(err, service.ErrBadMetricType) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			logger.Log.Error("Internal server error in GetMetric", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": service.ErrInternal.Error()})
		}
		return
	}

	responseDTO := modelToDTO(metricModel)

	responseBody, err := responseDTO.MarshalJSON()
	if err != nil {
		logger.Log.Error("Internal server error in GetJSON", zap.Error(err))
		c.String(http.StatusInternalServerError, service.ErrInternal.Error())
		return
	}

	c.Data(http.StatusOK, "application/json", responseBody)
}

func (h *MetricHandler) GetAll(c *gin.Context) {
	metrics, err := h.metricService.GetAll()
	if err != nil {
		logger.Log.Error("Internal server error in GetAll", zap.Error(err))
		c.String(http.StatusInternalServerError, "Failed to get metrics")
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")

	err = metricsTemplate.Execute(c.Writer, metrics)
	if err != nil {
		logger.Log.Error("Internal server error rendering template", zap.Error(err))
		c.String(http.StatusInternalServerError, "Failed to render page")
	}
}

func (h *MetricHandler) Update(c *gin.Context) {
	metricType := c.Param("type")
	metricName := c.Param("name")
	metricValue := c.Param("value")

	if strings.TrimSpace(metricName) == "" {
		c.String(http.StatusNotFound, "The metric name is missing")
		return
	}

	err := h.metricService.Update(metricType, metricName, metricValue)
	if err != nil {
		if errors.Is(err, service.ErrBadMetricType) || errors.Is(err, service.ErrBadMetricValue) {
			c.String(http.StatusBadRequest, err.Error())
		} else {
			logger.Log.Error("Internal server error in Update", zap.Error(err))
			c.String(http.StatusInternalServerError, service.ErrInternal.Error())
		}
		return
	}

	c.Status(http.StatusOK)
}

func (h *MetricHandler) UpdateJSON(c *gin.Context) {
	var dto MetricDTO

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	if err := dto.UnmarshalJSON(body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if strings.TrimSpace(dto.ID) == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "the metric name is missing"})
		return
	}

	metricModel, err := dtoToModel(dto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.metricService.UpdateJSON(metricModel)
	if err != nil {
		logger.Log.Error("Internal server error in UpdateJSON", zap.Error(err))
		c.String(http.StatusInternalServerError, service.ErrInternal.Error())
		return
	}

	c.Status(http.StatusOK)
}
