package agent

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/mailru/easyjson/jwriter"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"go.uber.org/zap"

	"github.com/iPatrushevSergey/metrics/internal/config"
	"github.com/iPatrushevSergey/metrics/internal/handler"
	"github.com/iPatrushevSergey/metrics/internal/logger"
	"github.com/iPatrushevSergey/metrics/internal/model"
	"github.com/iPatrushevSergey/metrics/internal/retry"
)

// CustomStats holds agent-specific metrics.
type CustomStats struct {
	PollCount   int64
	RandomValue float64
}

// GopsutilStats holds system metrics from gopsutil (memory, CPU).
type GopsutilStats struct {
	TotalMemory    float64
	FreeMemory     float64
	CPUutilization []float64
}

// Agent collects metrics and sends them to the server.
type Agent struct {
	config      config.AgentConfig
	client      *http.Client
	logger      logger.Logger
	mu          sync.RWMutex
	retryConfig retry.RetryConfig // Retry configuration for HTTP requests

	// Metrics
	memStats      runtime.MemStats
	customStats   CustomStats
	gopsutilStats GopsutilStats

	// Worker pool for batch requests
	batchJobs      chan []handler.MetricDTO
	results        chan error
	workersStarted bool
	workersWg      sync.WaitGroup
}

// NewAgent creates an Agent with the given config and logger.
func NewAgent(config config.AgentConfig, logger logger.Logger) (*Agent, error) {
	if config.RateLimit < 0 {
		return nil, fmt.Errorf("rate limit must be greater than or equal to 0, got: %d", config.RateLimit)
	}

	return &Agent{
		config:      config,
		client:      &http.Client{Timeout: 2 * time.Second},
		logger:      logger,
		retryConfig: DefaultRetryConfig(),
	}, nil
}

// Stop stops the agent's workers
func (a *Agent) Stop() {
	a.stopWorkers()
}

// PollMetrics collects go runtime and custom metrics
func (a *Agent) PollMetrics(ctx context.Context) {
	ticker := time.NewTicker(a.config.PollInterval)
	defer ticker.Stop()

	a.logger.Info("The metric collector is running")
	for {
		select {
		case <-ticker.C:
			a.mu.Lock()

			runtime.ReadMemStats(&a.memStats)
			a.customStats.PollCount++
			a.customStats.RandomValue = rand.Float64()

			a.mu.Unlock()
			a.logger.Debug("The metrics have been updated")
		case <-ctx.Done():
			a.logger.Info("The metric collector has been stopped")
			return
		}
	}
}

// PollGopsutilMetrics collects gopsutil metrics
func (a *Agent) PollGopsutilMetrics(ctx context.Context) {
	ticker := time.NewTicker(a.config.PollInterval)
	defer ticker.Stop()

	a.logger.Info("The gopsutil metric collector is running")
	for {
		select {
		case <-ticker.C:
			a.mu.Lock()

			v, err := mem.VirtualMemory()
			if err != nil {
				a.logger.Error("Error getting memory stats", zap.Error(err))
			} else {
				a.gopsutilStats.TotalMemory = float64(v.Total)
				a.gopsutilStats.FreeMemory = float64(v.Free)
			}

			percents, err := cpu.Percent(time.Second, true)
			if err != nil {
				a.logger.Error("Error getting CPU stats", zap.Error(err))
				a.gopsutilStats.CPUutilization = []float64{}
			} else {
				a.gopsutilStats.CPUutilization = percents
			}

			a.mu.Unlock()
			a.logger.Debug("The gopsutil metrics have been updated")
		case <-ctx.Done():
			a.logger.Info("The gopsutil metric collector has been stopped")
			return
		}
	}
}

// ReportMetrics report metrics
// Note: Ticker fires at scheduled times (every ReportInterval) regardless of whether ticks are read.
// If a tick is not read, it's buffered (buffer size = 1) and read immediately when select returns.
// If more than one tick is missed, extra ticks are lost.
func (a *Agent) ReportMetrics(ctx context.Context) {
	ticker := time.NewTicker(a.config.ReportInterval)
	defer ticker.Stop()

	a.logger.Info("The metrics sender is running")
	for {
		select {
		case <-ticker.C:
			a.logger.Debug("The beginning of sending metrics")
			go a.reportMetrics(ctx)
		case <-ctx.Done():
			a.logger.Info("The metrics sender has been stopped")
			return
		}
	}
}

// reportMetrics sends all metrics in batch mode
func (a *Agent) reportMetrics(ctx context.Context) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if duration > a.config.ReportInterval {
			a.logger.Warn(
				"Metrics sending took longer than report interval, possible channel blocking",
				zap.Duration("duration", duration),
				zap.Duration("report_interval", a.config.ReportInterval),
				zap.Duration("exceeded_by", duration-a.config.ReportInterval),
			)
		}
	}()

	if err := a.sendAllMetricsBatch(ctx); err != nil {
		if errors.Is(err, context.Canceled) {
			a.logger.Info("Sending metrics batch canceled")
			return
		}
		a.logger.Error("Error sending metrics batch", zap.Error(err))
	}
}

// startWorkers starts workers for processing batch requests
func (a *Agent) startWorkers(ctx context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.workersStarted {
		return
	}

	workerCount := a.config.RateLimit
	// Channel buffer size equals number of readers (workers) to prevent blocking
	a.batchJobs = make(chan []handler.MetricDTO, workerCount)
	// Channel buffer size equals number of writers (workers) to prevent blocking
	a.results = make(chan error, workerCount)

	for w := 1; w <= workerCount; w++ {
		a.workersWg.Add(1)
		go a.batchWorker(ctx, w, a.batchJobs, a.results, &a.workersWg)
	}

	go func() {
		for err := range a.results {
			if err != nil && !errors.Is(err, context.Canceled) {
				a.logger.Error("Error sending metrics batch", zap.Error(err))
			}
		}
	}()

	a.workersStarted = true
}

// stopWorkers stops the workers
func (a *Agent) stopWorkers() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.workersStarted {
		return
	}

	close(a.batchJobs)
	a.workersWg.Wait()
	close(a.results)
	a.workersStarted = false
}

// batchWorker processes batch requests from the jobs channel and sends results to the results channel
func (a *Agent) batchWorker(ctx context.Context, id int, batchJobs <-chan []handler.MetricDTO, results chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	for metrics := range batchJobs {
		select {
		case <-ctx.Done():
			results <- ctx.Err()
			return
		default:
			err := a.sendMetricsBatchRequest(ctx, metrics)
			if err != nil && !errors.Is(err, context.Canceled) {
				a.logger.Error("Batch worker error", zap.Int("worker_id", id), zap.Error(err))
				results <- err
			}
		}
	}
}

// sendMetricRequest sends a metric to the endpoint /update
func (a *Agent) sendMetricRequest(ctx context.Context, mType, mName string, mValue interface{}, workerID ...int) error {
	url := fmt.Sprintf("%s/update", a.config.Address)
	metricDTO := handler.MetricDTO{ID: mName, MType: mType}

	// Type definition
	switch model.MetricType(mType) {
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
	bodyBytes, err := metricDTO.MarshalJSON()
	if err != nil {
		return fmt.Errorf("error marshaling metric body: %w", err)
	}

	// Send request with retry logic
	response, err := sendRequestWithRetry(ctx, a.client, a.retryConfig, url, bodyBytes, a.config.Key)
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
			a.logger.Error("Error reading the response body", zap.Error(err))
		}
		return fmt.Errorf("failed status code: %s, body: %s", response.Status, string(body))
	}

	fields := []zap.Field{
		zap.String("metric", mName),
		zap.String("status", response.Status),
	}
	if len(workerID) > 0 && workerID[0] > 0 {
		fields = append(fields, zap.Int("worker", workerID[0]))
	}
	a.logger.Debug("Successfully sent metric", fields...)
	return nil
}

// sendAllMetricsBatch collects all metrics and sends them via worker pool or directly
func (a *Agent) sendAllMetricsBatch(ctx context.Context) error {
	select {
	case <-ctx.Done():
		a.logger.Info("The metrics sender has been stopped before batch sending")
		return ctx.Err()
	default:
	}

	a.mu.RLock()
	ms := a.memStats
	cs := a.customStats
	gs := a.gopsutilStats
	a.mu.RUnlock()

	gaugeMetrics := getGaugeMetrics(&ms, &cs, &gs)
	counterMetrics := getCounterMetrics(&ms, &cs)

	total := len(gaugeMetrics) + len(counterMetrics)
	if total == 0 {
		a.logger.Debug("No metrics to send in batch")
		return nil
	}

	metrics := make([]handler.MetricDTO, 0, total)

	for name, value := range gaugeMetrics {
		val := value
		metrics = append(metrics, handler.MetricDTO{
			ID:    name,
			MType: string(model.Gauge),
			Value: &val,
		})
	}

	for name, value := range counterMetrics {
		val := value
		metrics = append(metrics, handler.MetricDTO{
			ID:    name,
			MType: string(model.Counter),
			Delta: &val,
		})
	}

	// If rate limit is 0, send directly without worker pool
	if a.config.RateLimit == 0 {
		return a.sendMetricsBatchRequest(ctx, metrics)
	}

	// Use worker pool to limit concurrent batch requests
	a.startWorkers(ctx)

	// Try to send batch to worker pool with timeout to prevent goroutine accumulation
	select {
	case <-ctx.Done():
		return ctx.Err()
	case a.batchJobs <- metrics:
		// Batch request queued, worker will process it
		return nil
	case <-time.After(1 * time.Second):
		// Channel is full, log warning but don't block indefinitely
		// This prevents goroutine accumulation when server is very slow
		a.logger.Warn(
			"Batch jobs channel is full, dropping batch to prevent goroutine accumulation",
			zap.Int("rate_limit", a.config.RateLimit),
			zap.Int("batch_size", len(metrics)),
		)
		return fmt.Errorf("batch jobs channel is full, rate limit exceeded")
	}
}

// sendMetricsBatchRequest sends a batch of metrics to the endpoint /updates with gzip compression
func (a *Agent) sendMetricsBatchRequest(ctx context.Context, metrics []handler.MetricDTO) error {
	if len(metrics) == 0 {
		return nil
	}

	url := fmt.Sprintf("%s/updates", a.config.Address)

	// Request body formation
	var w jwriter.Writer
	w.RawByte('[')
	for i, m := range metrics {
		if i > 0 {
			w.RawByte(',')
		}
		m.MarshalEasyJSON(&w)
	}
	w.RawByte(']')

	bodyBytes := w.Buffer.BuildBytes()
	if w.Error != nil {
		return fmt.Errorf("error marshaling metrics batch body: %w", w.Error)
	}

	// Send request with retry logic
	response, err := sendRequestWithRetry(ctx, a.client, a.retryConfig, url, bodyBytes, a.config.Key)
	if err != nil {
		return fmt.Errorf("request sending error for batch: %w", err)
	}
	defer response.Body.Close()

	// Unpacking the response if necessary
	var reader io.ReadCloser
	switch response.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(response.Body)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader for batch response: %w", err)
		}
		defer reader.Close()
	default:
		reader = response.Body
	}

	// Response processing
	if response.StatusCode != http.StatusOK {
		body, err := io.ReadAll(reader)
		if err != nil {
			a.logger.Error("Error reading the batch response body", zap.Error(err))
		}
		return fmt.Errorf("failed batch status code: %s, body: %s", response.Status, string(body))
	}

	a.logger.Debug(
		"Successfully sent batch of metrics",
		zap.Int("count", len(metrics)),
		zap.String("status", response.Status),
	)
	return nil
}

func getGaugeMetrics(ms *runtime.MemStats, cs *CustomStats, gs *GopsutilStats) map[string]float64 {
	metrics := map[string]float64{
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
		"TotalMemory":   gs.TotalMemory,
		"FreeMemory":    gs.FreeMemory,
	}

	for i, cpuUtil := range gs.CPUutilization {
		metrics[fmt.Sprintf("CPUutilization%d", i+1)] = cpuUtil
	}

	return metrics
}

func getCounterMetrics(ms *runtime.MemStats, cs *CustomStats) map[string]int64 {
	return map[string]int64{
		"PollCount": cs.PollCount,
	}
}
