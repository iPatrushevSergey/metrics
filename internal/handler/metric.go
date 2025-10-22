package handler

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"html/template"

	"github.com/gin-gonic/gin"
	"github.com/iPatrushevSergey/metrics/internal/model"
	"github.com/iPatrushevSergey/metrics/internal/service"
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

type MetricHandler struct {
	metricService *service.MetricsService
}

func NewMetricHandler(s *service.MetricsService) *MetricHandler {
	return &MetricHandler{metricService: s}
}

func formatMetricValue(metric model.Metric) (string, error) {
	switch metric.MType {
	case model.Gauge:
		if metric.Value == nil {
			return "", fmt.Errorf("gauge value is nil")
		}
		return strconv.FormatFloat(*metric.Value, 'f', -1, 64), nil
	case model.Counter:
		if metric.Delta == nil {
			return "", fmt.Errorf("counter value is nil")
		}
		return strconv.FormatInt(*metric.Delta, 10), nil
	default:
		return "", fmt.Errorf("unknown metric MType: %s", metric.MType)
	}
}

func (h *MetricHandler) Get(c *gin.Context) {
	metricType := strings.ToLower(c.Param("type"))
	metricName := strings.TrimSpace(strings.ToLower(c.Param("name")))

	switch metricType {
	case model.Gauge, model.Counter:
		metric, err := h.metricService.Get(metricName)
		if err != nil {
			c.String(http.StatusNotFound, err.Error())
			return
		}

		value, err := formatMetricValue(metric)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		c.String(http.StatusOK, "%s", value)
	default:
		c.String(http.StatusBadRequest, "invalid metric type")
		return
	}
}

func (h *MetricHandler) GetAll(c *gin.Context) {
	metrics := h.metricService.GetAll()

	keys := make([]string, 0, len(metrics))
	for key := range metrics {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	data := struct {
		Metrics []templateData
	}{}

	for _, key := range keys {
		value, err := formatMetricValue(metrics[key])
		if err != nil {
			log.Printf("error formatting metric %s: %v", key, err)
			continue
		}
		data.Metrics = append(data.Metrics, templateData{Name: key, Value: value})
	}
	c.Header("Content-Type", "text/html; charset=utf-8")

	err := metricsTemplate.Execute(c.Writer, data)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to render page")
	}
}

func (h *MetricHandler) Update(c *gin.Context) {
	// Transport validation
	// contentType := c.GetHeader("Content-Type")
	// mediaType, params, err := mime.ParseMediaType(contentType)

	// if err != nil || mediaType != "text/plain" {
	// 	c.String(http.StatusUnsupportedMediaType, "Expected Content-Type: text/plain")
	// 	return
	// }

	// if charset := params["charset"]; charset != "" && charset != "utf-8" {
	// 	c.String(http.StatusUnsupportedMediaType, "Unsupported charset")
	// 	return
	// }

	metricType := strings.ToLower(c.Param("type"))
	metricName := strings.TrimSpace(strings.ToLower(c.Param("name")))
	metricValue := strings.ToLower(c.Param("value"))

	if metricName == "" {
		c.String(http.StatusNotFound, "The metric name is missing")
		return
	}

	// Format validation, parsing, update
	switch metricType {
	case model.Gauge:
		val, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid gauge value")
			return
		}
		err = h.metricService.Update(metricType, metricName, val)
		if err != nil {
			c.String(http.StatusInternalServerError, "failed to update metric")
			return
		}
	case model.Counter:
		val, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid counter value")
			return
		}
		err = h.metricService.Update(metricType, metricName, val)
		if err != nil {
			c.String(http.StatusInternalServerError, "failed to update metric")
			return
		}
	default:
		c.String(http.StatusBadRequest, "Invalid metric type")
		return
	}

	c.Status(http.StatusOK)
}
