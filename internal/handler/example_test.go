// Package handler_test contains examples for the handler package.
// These examples run with `go test ./internal/handler/ -run Example`.
// The "Run" button on pkg.go.dev does not work for internal packages.
package handler_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/iPatrushevSergey/metrics/internal/audit"
	"github.com/iPatrushevSergey/metrics/internal/handler"
	"github.com/iPatrushevSergey/metrics/internal/logger"
	"github.com/iPatrushevSergey/metrics/internal/repository/inmemory"
	"github.com/iPatrushevSergey/metrics/internal/service"
	"go.uber.org/zap"
)

func exampleRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	repo := inmemory.NewMemStorageMetricRepository()
	svc := service.NewMetricService(repo)
	log := logger.NewZapLoggerAdapter(zap.NewNop())
	h := handler.NewMetricHandler(svc, log, audit.NewPublisher(nil))
	r := gin.New()
	r.POST("/update/:type/:name/:value", h.Update)
	r.GET("/value/:type/:name", h.GetValue)
	r.POST("/update/", h.UpdateJSON)
	r.POST("/value/", h.GetJSON)
	r.GET("/", h.GetAll)
	r.GET("/ping", h.PingDB)
	return r
}

// Example demonstrates the metrics API: update a gauge via URL, then read it back.
func Example() {
	router := exampleRouter()

	// Update gauge via URL path
	req := httptest.NewRequest(http.MethodPost, "/update/gauge/alloc/1024.5", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Get value as plain text
	req2 := httptest.NewRequest(http.MethodGet, "/value/gauge/alloc", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	body, _ := io.ReadAll(w2.Result().Body)
	w2.Result().Body.Close()
	fmt.Println(string(body))
	// Output:
	// 1024.5
}

// ExampleMetricHandler_Update updates a metric via URL (POST /update/:type/:name/:value).
func ExampleMetricHandler_Update() {
	router := exampleRouter()

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/example_gauge/42.5", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	fmt.Println("status:", resp.StatusCode)
	// Output:
	// status: 200
}

// ExampleMetricHandler_GetValue returns a metric value as plain text (GET /value/:type/:name).
func ExampleMetricHandler_GetValue() {
	router := exampleRouter()

	// First update the metric
	upReq := httptest.NewRequest(http.MethodPost, "/update/gauge/alloc/1024.5", nil)
	upW := httptest.NewRecorder()
	router.ServeHTTP(upW, upReq)

	// Then get it
	req := httptest.NewRequest(http.MethodGet, "/value/gauge/alloc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	body, _ := io.ReadAll(w.Result().Body)
	w.Result().Body.Close()
	fmt.Println(string(body))
	// Output:
	// 1024.5
}

// ExampleMetricHandler_UpdateJSON updates a metric via JSON body (POST /update/).
func ExampleMetricHandler_UpdateJSON() {
	router := exampleRouter()

	body := []byte(`{"id":"cpu_usage","type":"gauge","value":0.95}`)
	req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Println("status:", w.Result().StatusCode)
	// Output:
	// status: 200
}

// ExampleMetricHandler_GetJSON returns a metric as JSON (POST /value/ with JSON body).
func ExampleMetricHandler_GetJSON() {
	router := exampleRouter()

	// Set a metric first
	upBody := []byte(`{"id":"memory","type":"gauge","value":512.0}`)
	upReq := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(upBody))
	upReq.Header.Set("Content-Type", "application/json")
	upW := httptest.NewRecorder()
	router.ServeHTTP(upW, upReq)

	// Get as JSON
	getBody := []byte(`{"id":"memory","type":"gauge"}`)
	getReq := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewReader(getBody))
	getReq.Header.Set("Content-Type", "application/json")
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	respBody, _ := io.ReadAll(getW.Result().Body)
	getW.Result().Body.Close()
	fmt.Println(string(respBody))
	// Output:
	// {"id":"memory","type":"gauge","value":512}
}
