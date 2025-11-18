package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/iPatrushevSergey/metrics/internal/config"
	"github.com/iPatrushevSergey/metrics/internal/handler"
	"github.com/iPatrushevSergey/metrics/internal/model"
)

type CustomStats struct {
	PollCount   int64
	RandomValue float64
}

type Agent struct {
	config config.AgentConfig
	client *http.Client
	mu     sync.RWMutex // The pattern "Critical Section"

	// Metrics
	memStats    runtime.MemStats
	customStats CustomStats
}

// The pattern "Constructor"
func NewAgent(config config.AgentConfig) *Agent {
	return &Agent{
		config: config,
		client: &http.Client{Timeout: 2 * time.Second},
	}
}

// The pattern "Worker"
func (a *Agent) PollMetrics(ctx context.Context) {
	ticker := time.NewTicker(a.config.PollInterval)
	defer ticker.Stop()

	log.Println("The metric collector is running")
	for {
		select {
		case <-ticker.C:
			a.mu.Lock()

			runtime.ReadMemStats(&a.memStats)
			a.customStats.PollCount++
			a.customStats.RandomValue = rand.Float64()

			a.mu.Unlock()
			log.Println("The metrics have been updated")
		case <-ctx.Done():
			log.Println("The metric collector has been stopped")
			return
		}
	}
}

// The pattern "Worker"
func (a *Agent) ReportMetrics(ctx context.Context) {
	ticker := time.NewTicker(a.config.ReportInterval)
	defer ticker.Stop()

	log.Println("The metrics sender is running")
	for {
		select {
		case <-ticker.C:
			log.Println("The beginning of sending metrics")
			a.sendAllMetrics(ctx)
		case <-ctx.Done():
			log.Println("The metrics sender has been stopped")
			return
		}
	}
}

func (a *Agent) sendAllMetrics(ctx context.Context) {
	// TODO: it makes sense to implement query bundling
	a.mu.RLock()
	ms := a.memStats
	cs := a.customStats
	a.mu.RUnlock()

	gaugeMetrics := getGaugeMetrics(&ms, &cs)
	for name, value := range gaugeMetrics {
		select {
		case <-ctx.Done():
			log.Println("The metrics sender has been stopped")
			return
		default:
		}

		if err := a.sendMetric(ctx, model.Gauge, name, value); err != nil {
			if errors.Is(err, context.Canceled) {
				log.Printf("Sending metric %s canceled\n", name)
				return
			}
			log.Printf("Error sending the metric gauge %s: %v\n", name, err)
		}
	}

	counterMetrics := getCounterMetrics(&ms, &cs)
	for name, value := range counterMetrics {
		select {
		case <-ctx.Done():
			log.Println("The metrics sender has been stopped")
			return
		default:
		}

		if err := a.sendMetric(ctx, model.Counter, name, value); err != nil {
			if errors.Is(err, context.Canceled) {
				log.Printf("Sending metric %s canceled\n", name)
				return
			}
			log.Printf("Error sending the metric counter %s: %v\n", name, err)
		}
	}
}

func (a *Agent) sendMetric(ctx context.Context, mType, mName string, mValue interface{}) error {
	url := fmt.Sprintf("%s/update", a.config.Address)
	metricDTO := handler.MetricDTO{ID: mName, MType: mType}

	// Type definition
	switch mType {
	case model.Counter:
		val, ok := mValue.(int64)
		if !ok {
			return fmt.Errorf("invalid value type for Counter metric %s", mName)
		}
		metricDTO.Delta = &val
	case model.Gauge:
		val, ok := mValue.(float64)
		if !ok {
			return fmt.Errorf("invalid value type for Gauge metric %s", mName)
		}
		metricDTO.Value = &val
	default:
		return fmt.Errorf("unknown metric type: %s", mType)
	}

	// Request body formation
	bodyBytes, err := json.Marshal(metricDTO)
	if err != nil {
		return fmt.Errorf("error marshaling metric body: %w", err)
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(bodyBytes); err != nil {
		return fmt.Errorf("error compressing body: %w", err)
	}
	if err := gz.Close(); err != nil {
		return fmt.Errorf("error closing gzip writer: %w", err)
	}

	// Request formation
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		return fmt.Errorf("request creation error: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Encoding", "gzip")
	req.Header.Add("Accept-Encoding", "gzip")

	// Send request
	response, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("request sending error: %w", err)
	}
	defer response.Body.Close()

	// Unpacking
	var reader io.ReadCloser
	switch response.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(response.Body)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader for response: %w", err)
		}
		defer reader.Close()
	default:
		reader = response.Body
	}

	// Response processing
	if response.StatusCode != http.StatusOK {
		body, err := io.ReadAll(reader)
		if err != nil {
			log.Printf("Error reading the response body: %v\n", err)
		}
		return fmt.Errorf("failed status code: %s, body: %s", response.Status, string(body))
	}

	log.Printf("Successfully sent: %s. Status: %s\n", mName, response.Status)
	return nil
}

func getGaugeMetrics(ms *runtime.MemStats, cs *CustomStats) map[string]float64 {
	return map[string]float64{
		"Alloc":         float64(ms.Alloc),
		"BuckHashSys":   float64(ms.BuckHashSys),
		"Frees":         float64(ms.Frees),
		"GCCPUFraction": ms.GCCPUFraction,
		"GCSys":         float64(ms.GCSys),
		"HeapAlloc":     float64(ms.HeapAlloc),
		"HeapIdle":      float64(ms.HeapIdle),
		"HeapInuse":     float64(ms.HeapInuse),
		"HeapObjects":   float64(ms.HeapObjects),
		"HeapReleased":  float64(ms.HeapReleased),
		"HeapSys":       float64(ms.HeapSys),
		"LastGC":        float64(ms.LastGC),
		"Lookups":       float64(ms.Lookups),
		"MCacheInuse":   float64(ms.MCacheInuse),
		"MCacheSys":     float64(ms.MCacheSys),
		"MSpanInuse":    float64(ms.MSpanInuse),
		"MSpanSys":      float64(ms.MSpanSys),
		"Mallocs":       float64(ms.Mallocs),
		"NextGC":        float64(ms.NextGC),
		"NumForcedGC":   float64(ms.NumForcedGC),
		"NumGC":         float64(ms.NumGC),
		"OtherSys":      float64(ms.OtherSys),
		"PauseTotalNs":  float64(ms.PauseTotalNs),
		"StackInuse":    float64(ms.StackInuse),
		"StackSys":      float64(ms.StackSys),
		"Sys":           float64(ms.Sys),
		"TotalAlloc":    float64(ms.TotalAlloc),
		"RandomValue":   cs.RandomValue,
	}
}

func getCounterMetrics(ms *runtime.MemStats, cs *CustomStats) map[string]int64 {
	return map[string]int64{
		"PollCount": cs.PollCount,
	}
}
