package agent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/iPatrushevSergey/metrics/internal/config"
	"github.com/iPatrushevSergey/metrics/internal/logger"
	"github.com/stretchr/testify/require"
)

func TestNewAgent_negativeRateLimit(t *testing.T) {
	_, err := NewAgent(config.AgentConfig{RateLimit: -1}, logger.NewZapLoggerAdapter(zap.NewNop()))
	require.Error(t, err)
}

func TestStop_withoutWorkers(t *testing.T) {
	a, err := NewAgent(config.AgentConfig{RateLimit: 0}, logger.NewZapLoggerAdapter(zap.NewNop()))
	require.NoError(t, err)
	a.Stop() // stopWorkers early return
}

func TestSendAllMetricsBatch_ctxCanceled(t *testing.T) {
	a, err := NewAgent(config.AgentConfig{RateLimit: 0}, logger.NewZapLoggerAdapter(zap.NewNop()))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = a.sendAllMetricsBatch(ctx)
	require.ErrorIs(t, err, context.Canceled)
}

func newUpdatesOKServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/updates" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
}

func TestSendAllMetricsBatch_directOK(t *testing.T) {
	srv := newUpdatesOKServer(t)
	t.Cleanup(srv.Close)

	a, err := NewAgent(config.AgentConfig{
		Address:   srv.URL,
		RateLimit: 0,
	}, logger.NewZapLoggerAdapter(zap.NewNop()))
	require.NoError(t, err)

	a.mu.Lock()
	runtime.ReadMemStats(&a.memStats)
	a.customStats.PollCount = 1
	a.mu.Unlock()

	require.NoError(t, a.sendAllMetricsBatch(context.Background()))
}

func TestSendAllMetricsBatch_workerPool_andStop(t *testing.T) {
	srv := newUpdatesOKServer(t)
	t.Cleanup(srv.Close)

	a, err := NewAgent(config.AgentConfig{
		Address:        srv.URL,
		RateLimit:      1,
		PollInterval:   time.Hour,
		ReportInterval: time.Hour,
	}, logger.NewZapLoggerAdapter(zap.NewNop()))
	require.NoError(t, err)

	a.mu.Lock()
	runtime.ReadMemStats(&a.memStats)
	a.customStats.PollCount = 1
	a.mu.Unlock()

	ctx := context.Background()
	require.NoError(t, a.sendAllMetricsBatch(ctx))

	time.Sleep(150 * time.Millisecond)
	a.Stop()
}

func TestSendMetricsBatchRequest_emptySlice(t *testing.T) {
	a, err := NewAgent(config.AgentConfig{Address: "http://localhost", RateLimit: 0}, logger.NewZapLoggerAdapter(zap.NewNop()))
	require.NoError(t, err)
	require.NoError(t, a.sendMetricsBatchRequest(context.Background(), nil))
}

func TestReportMetrics_stopsWithContext(t *testing.T) {
	srv := newUpdatesOKServer(t)
	t.Cleanup(srv.Close)

	a, err := NewAgent(config.AgentConfig{
		Address:        srv.URL,
		ReportInterval: 25 * time.Millisecond,
		RateLimit:      0,
	}, logger.NewZapLoggerAdapter(zap.NewNop()))
	require.NoError(t, err)

	a.mu.Lock()
	runtime.ReadMemStats(&a.memStats)
	a.customStats.PollCount = 1
	a.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.ReportMetrics(ctx, ctx)
	}()

	time.Sleep(60 * time.Millisecond)
	cancel()
	wg.Wait()
	a.WaitSendsDone()
}
