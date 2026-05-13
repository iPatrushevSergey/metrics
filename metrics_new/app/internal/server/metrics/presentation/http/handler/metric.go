package handler

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application"
	appdto "github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/port"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/presentation/factory"
	httpdto "github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/presentation/http/dto"
)

const metricsHTMLTemplate = `
	<html>
	<head><title>Metrics</title></head>
	<body>
		<h1>All metrics</h1>
		<ul>
			{{ range .Metrics }}
			<li><b>{{ .MetricID }}:</b> {{ .MetricValue }}</li>
			{{ end }}
		</ul>
	</body>
</html>
`

var metricsTemplate = template.Must(template.New("metrics").Parse(metricsHTMLTemplate))

// MetricHandler serves the metrics API.
type MetricHandler struct {
	useCases factory.UseCaseFactory
	log      port.Logger
}

// NewMetricHandler constructs the handler with injected use cases.
func NewMetricHandler(uc factory.UseCaseFactory, log port.Logger) *MetricHandler {
	return &MetricHandler{useCases: uc, log: log}
}

// GetValue returns the metric value.
func (h *MetricHandler) GetValue(c *gin.Context) {
	ctx := c.Request.Context()

	inDTO := appdto.GetMetricValueInput{
		MType: strings.ToLower(strings.TrimSpace(c.Param("type"))),
		ID:    strings.TrimSpace(c.Param("name")),
	}

	metricValue, err := h.useCases.GetMetricValueUseCase().Execute(ctx, inDTO)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrNotFound):
			c.String(http.StatusNotFound, err.Error())
			return
		case errors.Is(err, application.ErrBadMetricType):
			c.String(http.StatusBadRequest, err.Error())
			return
		default:
			h.log.Error("get metric value failed", "error", err)
			c.String(http.StatusInternalServerError, application.ErrInternal.Error())
			return
		}
	}
	c.String(http.StatusOK, "%s", metricValue)
}

// GetJSON returns a single metric as JSON.
func (h *MetricHandler) GetJSON(c *gin.Context) {
	ctx := c.Request.Context()

	var reqDTO httpdto.MetricRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&reqDTO); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	inDTO := appdto.GetMetricInput{
		MType: strings.ToLower(strings.TrimSpace(reqDTO.MType)),
		ID:    strings.TrimSpace(reqDTO.ID),
	}

	metric, err := h.useCases.GetMetricUseCase().Execute(ctx, inDTO)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		case errors.Is(err, application.ErrBadMetricType), errors.Is(err, application.ErrBadMetricValue):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		default:
			h.log.Error("get metric json failed", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": application.ErrInternal.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, httpdto.MetricResponse{
		ID:    metric.ID,
		MType: metric.MType,
		Delta: metric.Delta,
		Value: metric.Value,
	})
}

// GetAll returns all metrics as HTML.
func (h *MetricHandler) GetAll(c *gin.Context) {
	ctx := c.Request.Context()

	metricsForDisplay, err := h.useCases.GetAllMetricsUseCase().Execute(ctx, struct{}{})
	if err != nil {
		h.log.Error("get all metrics failed", "error", err)
		c.String(http.StatusInternalServerError, application.ErrInternal.Error())
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := metricsTemplate.Execute(c.Writer, gin.H{"Metrics": metricsForDisplay}); err != nil {
		h.log.Error("render metrics template", "error", err)
	}
}

// Update updates or creates a metric from the URL parameters.
func (h *MetricHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()

	inDTO := appdto.UpdateMetricInput{
		MType: strings.ToLower(strings.TrimSpace(c.Param("type"))),
		ID:    strings.TrimSpace(c.Param("name")),
		Value: strings.TrimSpace(c.Param("value")),
	}

	if strings.TrimSpace(inDTO.ID) == "" {
		c.String(http.StatusNotFound, "The metric name is missing")
		return
	}

	if _, err := h.useCases.UpdateMetricUseCase().Execute(ctx, inDTO); err != nil {
		switch {
		case errors.Is(err, application.ErrBadMetricValue), errors.Is(err, application.ErrBadMetricType):
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		h.log.Error("update metric failed", "error", err)
		c.String(http.StatusInternalServerError, application.ErrInternal.Error())
		return
	}
	c.Status(http.StatusOK)
}

// UpdateJSON updates or creates a metric from the JSON request body
func (h *MetricHandler) UpdateJSON(c *gin.Context) {
	ctx := c.Request.Context()

	var reqDTO httpdto.MetricRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&reqDTO); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if strings.TrimSpace(reqDTO.ID) == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "the metric name is missing"})
		return
	}

	inDTO := appdto.UpsertMetricInput{
		ID:    strings.TrimSpace(reqDTO.ID),
		MType: reqDTO.MType,
		Delta: reqDTO.Delta,
		Value: reqDTO.Value,
	}

	if _, err := h.useCases.UpsertMetricUseCase().Execute(ctx, inDTO); err != nil {
		switch {
		case errors.Is(err, application.ErrBadMetricValue), errors.Is(err, application.ErrBadMetricType):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		h.log.Error("update json metric failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": application.ErrInternal.Error()})
		return
	}
	c.Status(http.StatusOK)
}

// UpdatesJSON updates or creates a list of metrics from the JSON request body.
func (h *MetricHandler) UpdatesJSON(c *gin.Context) {
	ctx := c.Request.Context()

	var reqDTOs []httpdto.MetricRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&reqDTOs); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	inDTOs := make([]appdto.UpsertMetricInput, 0, len(reqDTOs))
	for _, reqDTO := range reqDTOs {
		if strings.TrimSpace(reqDTO.ID) == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": "the metric name is missing"})
			return
		}
		inDTOs = append(inDTOs, appdto.UpsertMetricInput{
			ID:    strings.TrimSpace(reqDTO.ID),
			MType: reqDTO.MType,
			Delta: reqDTO.Delta,
			Value: reqDTO.Value,
		})
	}

	if _, err := h.useCases.UpsertMetricsBatchUseCase().Execute(ctx, appdto.UpsertMetricsBatchInput{Metrics: inDTOs}); err != nil {
		switch {
		case errors.Is(err, application.ErrBadMetricValue), errors.Is(err, application.ErrBadMetricType):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		h.log.Error("update json batch metrics failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": application.ErrInternal.Error()})
		return
	}
	c.Status(http.StatusOK)
}

// PingDB checks the availability of the database.
func (h *MetricHandler) PingDB(c *gin.Context) {
	ctx := c.Request.Context()

	if _, err := h.useCases.PingDBUseCase().Execute(ctx, struct{}{}); err != nil {
		h.log.Error("ping db failed", "error", err)
		c.String(http.StatusInternalServerError, application.ErrInternal.Error())
		return
	}
	c.Status(http.StatusOK)
}
