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

type CustomStats struct {
	PollCount   int64
	RandomValue float64
}

type GopsutilStats struct {
	TotalMemory    float64
	FreeMemory     float64
	CPUutilization []float64
}

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

	// Worker pool
	jobs           chan MetricTask
	results        chan error
	workersStarted bool
	workersWg      sync.WaitGroup
}

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
func (a *Agent) ReportMetrics(ctx context.Context) {
	ticker := time.NewTicker(a.config.ReportInterval)
	defer ticker.Stop()

	a.logger.Info("The metrics sender is running")
	for {
		select {
		case <-ticker.C:
			a.logger.Debug("The beginning of sending metrics")
			a.reportMetrics(ctx)
		case <-ctx.Done():
			a.logger.Info("The metrics sender has been stopped")
			return
		}
	}
}

// reportMetrics selects the mode of sending metrics: single or batch
func (a *Agent) reportMetrics(ctx context.Context) {
	if a.config.UseBatchMode {
		a.logger.Debug("Sending metrics in batch mode")
		if err := a.sendAllMetricsBatch(ctx); err != nil {
			if errors.Is(err, context.Canceled) {
				a.logger.Info("Sending metrics batch canceled")
				return
			}
			a.logger.Error("Error sending metrics batch", zap.Error(err))
		}
		return
	}

	a.sendAllMetrics(ctx)
}

// MetricTask represents a task for sending a single metric
type MetricTask struct {
	MType  string
	MName  string
	MValue interface{}
}

// startWorkers starting workers
func (a *Agent) startWorkers(ctx context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.workersStarted {
		return
	}

	workerCount := a.config.RateLimit
	a.jobs = make(chan MetricTask, 100)
	a.results = make(chan error, 100)

	for w := 1; w <= workerCount; w++ {
		a.workersWg.Add(1)
		go a.worker(ctx, w, a.jobs, a.results, &a.workersWg)
	}

	go func() {
		for err := range a.results {
			if err != nil && !errors.Is(err, context.Canceled) {
				a.logger.Error("Error sending metric", zap.Error(err))
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

	close(a.jobs)
	a.workersWg.Wait()
	close(a.results)
	a.workersStarted = false
}

// sendAllMetrics
func (a *Agent) sendAllMetrics(ctx context.Context) {
	a.mu.RLock()
	ms := a.memStats
	cs := a.customStats
	gs := a.gopsutilStats
	a.mu.RUnlock()

	gaugeMetrics := getGaugeMetrics(&ms, &cs, &gs)
	counterMetrics := getCounterMetrics(&ms, &cs)

	workerCount := a.config.RateLimit

	if workerCount == 0 {
		a.sendAllMetricsSequential(ctx, gaugeMetrics, counterMetrics)
		return
	}

	a.startWorkers(ctx)

	for name, value := range gaugeMetrics {
		select {
		case <-ctx.Done():
			return
		default:
			select {
			case a.jobs <- MetricTask{MType: model.Gauge, MName: name, MValue: value}:
			case <-ctx.Done():
				return
			}
		}
	}

	for name, value := range counterMetrics {
		select {
		case <-ctx.Done():
			return
		default:
			select {
			case a.jobs <- MetricTask{MType: model.Counter, MName: name, MValue: value}:
			case <-ctx.Done():
				return
			}
		}
	}
}

// sendAllMetricsSequential sends metrics sequentially
func (a *Agent) sendAllMetricsSequential(ctx context.Context, gaugeMetrics map[string]float64, counterMetrics map[string]int64) {
	for name, value := range gaugeMetrics {
		select {
		case <-ctx.Done():
			a.logger.Info("The metrics sender has been stopped")
			return
		default:
		}

		if err := a.sendMetricRequest(ctx, model.Gauge, name, value); err != nil {
			if errors.Is(err, context.Canceled) {
				a.logger.Info("Sending metric canceled", zap.String("metric", name))
				return
			}
			a.logger.Error(
				"Error sending the metric gauge",
				zap.String("metric", name),
				zap.String("type", "gauge"),
				zap.Error(err),
			)
		}
	}

	for name, value := range counterMetrics {
		select {
		case <-ctx.Done():
			a.logger.Info("The metrics sender has been stopped")
			return
		default:
		}

		if err := a.sendMetricRequest(ctx, model.Counter, name, value); err != nil {
			if errors.Is(err, context.Canceled) {
				a.logger.Info("Sending metric canceled", zap.String("metric", name))
				return
			}
			a.logger.Error(
				"Error sending the metric counter",
				zap.String("metric", name),
				zap.String("type", "counter"),
				zap.Error(err),
			)
		}
	}
}

// worker processes tasks from the jobs channel and sends results to the results channel
func (a *Agent) worker(ctx context.Context, id int, jobs <-chan MetricTask, results chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		select {
		case <-ctx.Done():
			results <- ctx.Err()
			return
		default:
			err := a.sendMetricRequest(ctx, job.MType, job.MName, job.MValue, id)
			results <- err
		}
	}
}

// sendMetricRequest sends a metric to the endpoint /update
func (a *Agent) sendMetricRequest(ctx context.Context, mType, mName string, mValue interface{}, workerID ...int) error {
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

// sendAllMetricsBatch collects all metrics and sends them in one batch request
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
			MType: model.Gauge,
			Value: &val,
		})
	}

	for name, value := range counterMetrics {
		val := value
		metrics = append(metrics, handler.MetricDTO{
			ID:    name,
			MType: model.Counter,
			Delta: &val,
		})
	}

	return a.sendMetricsBatchRequest(ctx, metrics)
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
