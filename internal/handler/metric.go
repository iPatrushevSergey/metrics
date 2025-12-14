package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

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

type templateData struct {
	Name  string
	Value string
}

type responseMetrics struct {
	Metrics []templateData
}

type MetricHandler struct {
	metricService *service.MetricsService
	logger        logger.Logger
}

func NewMetricHandler(s *service.MetricsService, l logger.Logger) *MetricHandler {
	return &MetricHandler{
		metricService: s,
		logger:        l,
	}
}

// GetValue возвращает значение метрики в виде строки
func (h *MetricHandler) GetValue(c *gin.Context) {
	ctx := c.Request.Context()
	metricType := strings.ToLower(strings.TrimSpace(c.Param("type")))
	metricName := strings.TrimSpace(c.Param("name"))

	metricVal, err := h.metricService.GetValue(ctx, metricType, metricName)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.String(http.StatusNotFound, err.Error())
		} else if errors.Is(err, service.ErrBadMetricType) {
			c.String(http.StatusBadRequest, err.Error())
		} else {
			h.logger.Error("Internal server error in GetValue", zap.Error(err))
			c.String(http.StatusInternalServerError, service.ErrInternal.Error())
		}
		return
	}

	c.String(http.StatusOK, "%s", metricVal)
}

// GetJSON возвращает метрику в формате JSON
func (h *MetricHandler) GetJSON(c *gin.Context) {
	ctx := c.Request.Context()
	var dto MetricDTO

	if err := json.NewDecoder(c.Request.Body).Decode(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Нормализация данных перед передачей в service
	metricType := strings.ToLower(strings.TrimSpace(dto.MType))
	metricName := strings.TrimSpace(dto.ID)

	metricModel, err := h.metricService.GetMetric(ctx, metricType, metricName)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if errors.Is(err, service.ErrBadMetricType) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			h.logger.Error("Internal server error in GetMetric", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": service.ErrInternal.Error()})
		}
		return
	}

	responseDTO := modelToDTO(metricModel)
	c.JSON(http.StatusOK, responseDTO)
}

// GetAll возвращает все метрики в формате HTML
func (h *MetricHandler) GetAll(c *gin.Context) {
	ctx := c.Request.Context()

	metrics, err := h.metricService.GetAll(ctx)
	if err != nil {
		h.logger.Error("Internal server error in GetAll", zap.Error(err))
		c.String(http.StatusInternalServerError, "Failed to get metrics")
		return
	}

	keys := make([]string, 0, len(metrics))
	for key := range metrics {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	data := responseMetrics{}
	for _, key := range keys {
		value, err := h.metricService.FormatMetric(metrics[key])
		if err != nil {
			h.logger.Error("Failed to format metric for display", zap.String("metric", key), zap.Error(err))
			continue
		}
		data.Metrics = append(data.Metrics, templateData{Name: key, Value: value})
	}

	c.Header("Content-Type", "text/html; charset=utf-8")

	if err = metricsTemplate.Execute(c.Writer, data); err != nil {
		h.logger.Error("Internal server error rendering template", zap.Error(err))
		c.String(http.StatusInternalServerError, "Failed to render page")
	}
}

// Update обновляет или создает метрику из URL параметров
func (h *MetricHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	metricType := strings.ToLower(strings.TrimSpace(c.Param("type")))
	metricName := strings.TrimSpace(c.Param("name"))
	metricValue := strings.TrimSpace(c.Param("value"))

	if metricName == "" {
		c.String(http.StatusNotFound, "The metric name is missing")
		return
	}

	err := h.metricService.Update(ctx, metricType, metricName, metricValue)
	if err != nil {
		if errors.Is(err, service.ErrBadMetricType) || errors.Is(err, service.ErrBadMetricValue) {
			c.String(http.StatusBadRequest, err.Error())
		} else {
			h.logger.Error("Internal server error in Update", zap.Error(err))
			c.String(http.StatusInternalServerError, service.ErrInternal.Error())
		}
		return
	}

	c.Status(http.StatusOK)
}

// UpdateJSON обновляет или создает метрику из JSON тела запроса
func (h *MetricHandler) UpdateJSON(c *gin.Context) {
	ctx := c.Request.Context()
	var dto MetricDTO

	if err := json.NewDecoder(c.Request.Body).Decode(&dto); err != nil {
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

	if err = h.metricService.UpdateJSON(ctx, metricModel); err != nil {
		if errors.Is(err, service.ErrBadMetricValue) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			h.logger.Error("Internal server error in UpdateJSON", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": service.ErrInternal.Error()})
		}
		return
	}

	c.Status(http.StatusOK)
}

const (
	// pingDBTimeout таймаут для проверки доступности базы данных
	pingDBTimeout = 1 * time.Second
)

// PingDB проверяет доступность базы данных
func (h *MetricHandler) PingDB(c *gin.Context) {
	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, pingDBTimeout)
	defer cancel()

	if err := h.metricService.PingDB(ctx); err != nil {
		h.logger.Error("Database ping failed", zap.Error(err))
		c.String(http.StatusInternalServerError, service.ErrInternal.Error())
		return
	}
	c.Status(http.StatusOK)
}
