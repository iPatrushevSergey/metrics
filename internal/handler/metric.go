package handler

import (
	"mime"
	"net/http"
	"strconv"
	"strings"

	"github.com/iPatrushevSergey/metrics/internal/model"
	"github.com/iPatrushevSergey/metrics/internal/service"
)

type MetricHandler struct {
	metricService *service.MetricsService
}

func NewMetricHandler(s *service.MetricsService) *MetricHandler {
	return &MetricHandler{metricService: s}
}

func (h *MetricHandler) Update(w http.ResponseWriter, r *http.Request) {
	// Transport validation
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	contentType := r.Header.Get("Content-Type")
	mediaType, params, err := mime.ParseMediaType(contentType)

	if err != nil || mediaType != "text/plain" {
		http.Error(w, "Expected Content-Type: text/plain", http.StatusUnsupportedMediaType)
		return
	}

	if charset := params["charset"]; charset != "" && charset != "utf-8" {
		http.Error(w, "Unsupported charset", http.StatusUnsupportedMediaType)
		return
	}

	metricType := strings.ToLower(r.PathValue("type"))
	metricName := strings.ToLower(r.PathValue("name"))
	metricValue := strings.ToLower(r.PathValue("value"))

	if metricName == "" {
		http.Error(w, "The metric name is missing", http.StatusNotFound)
		return
	}

	// Format validation, parsing, update
	switch metricType {
	case model.Gauge:
		val, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(w, "Invalid gauge value", http.StatusBadRequest)
			return
		}
		h.metricService.Update(metricType, metricName, val)
	case model.Counter:
		val, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
			return
		}
		h.metricService.Update(metricType, metricName, val)
	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
