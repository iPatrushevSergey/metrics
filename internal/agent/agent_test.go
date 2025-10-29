package agent

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/iPatrushevSergey/metrics/internal/config"
	"github.com/iPatrushevSergey/metrics/internal/model"
)

func TestNewAgent(t *testing.T) {
	expectedConfig := config.AgentConfig{
		PollInterval:   3,
		ReportInterval: 6,
		Address:        "http//:127.0.0.1:8080",
	}

	agent := NewAgent(expectedConfig)

	if agent == nil {
		t.Fatalf("NewAgent returned nil, expected *Agent")
	}

	if agent.config != expectedConfig {
		t.Errorf("NewAgent config mismatch. Got %+v, expected %+v", agent.config, expectedConfig)
	}

	if agent.client == nil {
		t.Errorf("NewAgent client is nil, expected *http.Client")
	}

	expectedTimeout := 2 * time.Second
	if agent.client.Timeout != expectedTimeout {
		t.Errorf("NewAgent timeout mismatch. Got %v, expected %v", agent.client.Timeout, expectedTimeout)
	}

	if agent.memStats.Alloc != 0 {
		t.Errorf("NewAgent memStats.Alloc mismatch. Got %v, expected 0 (zero value)", agent.memStats.Alloc)
	}

	if agent.customStats.RandomValue != 0 {
		t.Errorf("NewAgent customStats.PollCount mismatch. Got %v, expected 0 (zero value)", agent.customStats.PollCount)
	}
}

func TestPollMetrics(t *testing.T) {
	testPollInterval := 10 * time.Millisecond
	testConfig := config.AgentConfig{
		PollInterval: testPollInterval,
	}
	agent := NewAgent(testConfig)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		agent.PollMetrics(ctx)
	}()

	wg.Wait()

	if agent.customStats.PollCount <= 0 {
		t.Errorf("PollCount has not increased. PollCounter got: %d, expected > 0", agent.customStats.PollCount)
	}

	if agent.customStats.RandomValue == 0 {
		t.Errorf("RandomValue has not increased. RandomValue got: %f, expected > 0", agent.customStats.RandomValue)
	}

	if agent.memStats.Alloc == 0 {
		t.Errorf("Alloc has not increased. Alloc got: %d, expected > 0", agent.memStats.Alloc)
	}
}

func TestSendMetric(t *testing.T) {
	type test struct {
		name          string
		metricType    string
		metricName    string
		metricValue   string
		wantCode      int
		wantErrorBody string
		wantError     bool
	}

	tests := []test{
		{
			name:          "Success Gauge 200",
			metricType:    model.Gauge,
			metricName:    "Alloc",
			metricValue:   "100.1",
			wantCode:      http.StatusOK,
			wantErrorBody: "",
			wantError:     false,
		},
		{
			name:          "Success Counter 200",
			metricType:    model.Counter,
			metricName:    "PollCount",
			metricValue:   "13",
			wantCode:      http.StatusOK,
			wantErrorBody: "",
			wantError:     false,
		},
		{
			name:          "Fail Gauge 404 NotFound",
			metricType:    model.Gauge,
			metricName:    "UnknownMetric",
			metricValue:   "1.0",
			wantCode:      http.StatusNotFound,
			wantErrorBody: "Metric not found on server",
			wantError:     true,
		},
		{
			name:          "Fail 400 BadRequest",
			metricType:    model.Counter,
			metricName:    "PollCount",
			metricValue:   "invalid_value",
			wantCode:      http.StatusBadRequest,
			wantErrorBody: "Invalid metric value format",
			wantError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := fmt.Sprintf("/update/%s/%s/%s", tt.metricType, tt.metricName, tt.metricValue)

				if r.URL.Path != expectedPath {
					t.Errorf("Handler received wrong path. Got: %s, expected: %s", r.URL.Path, expectedPath)
				}

				if r.Method != http.MethodPost {
					t.Errorf("Handler received wrong method. Got: %s, expected: POST", r.Method)
				}

				if r.Header.Get("Content-Type") != "text/plain" {
					t.Errorf("Handler received wrong Content-Type. Got: %s, expected: text/plain", r.Header.Get("Content-Type"))
				}

				w.WriteHeader(tt.wantCode)
				if tt.wantErrorBody != "" {
					io.WriteString(w, tt.wantErrorBody)
				}
			}))
			defer ts.Close()

			agent := NewAgent(config.AgentConfig{Address: ts.URL})
			err := agent.sendMetric(context.Background(), tt.metricType, tt.metricName, tt.metricValue)

			if tt.wantError {
				if err == nil {
					t.Fatalf("Expected an error but got nil")
				}

				expectedStatusText := http.StatusText(tt.wantCode)
				expectedErrorSubstring := fmt.Sprintf("%d %s", tt.wantCode, expectedStatusText)

				if !strings.Contains(err.Error(), expectedErrorSubstring) {
					t.Errorf("Error message mismatch. Got: %s\nExpected: %s", err.Error(), expectedErrorSubstring)
				}
				if tt.wantErrorBody != "" && !strings.Contains(err.Error(), tt.wantErrorBody) {
					t.Errorf("Error message missing body. Got: %s\nExpected: %s", err.Error(), tt.wantErrorBody)
				}
			} else {
				if err != nil {
					t.Fatalf("Expected nil error for success, but got: %v", err)
				}
			}
		})
	}
}

func TestGetGuaugeMetrics(t *testing.T) {
	const mockAlloc = 1024
	const mockGCCPUFraction = 0.123
	const mockRandomValue = 42.5

	ms := runtime.MemStats{
		Alloc:         mockAlloc,
		GCCPUFraction: mockGCCPUFraction,
	}

	cs := CustomStats{
		RandomValue: mockRandomValue,
	}

	metrics := getGaugeMetrics(&ms, &cs)

	if len(metrics) != 28 {
		t.Errorf("getGaugeMetrics returned %d metrics, expected 28", len(metrics))
	}

	tests := map[string]float64{
		"Alloc":         float64(mockAlloc),
		"GCCPUFraction": mockGCCPUFraction,
		"RandomValue":   mockRandomValue,
	}

	for name, expected := range tests {
		actual, ok := metrics[name]
		if !ok {
			t.Errorf("Metric %s not found", name)
		}
		if actual < expected-0.0001 || actual > expected+0.0001 {
			t.Errorf("Metric %s mismatch. Got %f, expected %f", name, actual, expected)
		}
	}
}

func TestGetCounterMetrics(t *testing.T) {
	const mockPollCount = 10

	var ms runtime.MemStats

	cs := CustomStats{
		PollCount: mockPollCount,
	}

	metrics := getCounterMetrics(&ms, &cs)

	if len(metrics) != 1 {
		t.Errorf("getCounterMetrics returned %d metrics, expected 1", len(metrics))
	}

	tests := map[string]int64{
		"PollCount": int64(mockPollCount),
	}

	for name, expected := range tests {
		actual, ok := metrics[name]
		if !ok {
			t.Errorf("Metric %s not found", name)
		}
		if actual != expected {
			t.Errorf("Metric %s mismatch. Got %d, expected %d", name, actual, expected)
		}
	}
}
