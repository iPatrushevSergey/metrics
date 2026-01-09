package middleware

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"github.com/iPatrushevSergey/metrics/internal/logger"
	"github.com/stretchr/testify/require"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	testLogger := logger.NewZapLoggerAdapter(zap.NewNop())
	router.Use(GzipGinMiddleware(testLogger))

	router.POST("/test-json", func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad body"})
			return
		}

		responseMap := gin.H{"received_body": string(body)}
		responseBody, err := json.Marshal(responseMap)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "marshal error"})
			return
		}

		c.Data(http.StatusOK, "application/json", responseBody)
	})

	router.GET("/test-html", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<h1>Hello</h1>"))
	})

	router.GET("/test-text", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/plain", []byte("No compression"))
	})

	return router
}

func TestGzipGinMiddleware(t *testing.T) {
	router := setupRouter()

	t.Run("Client sends gzipped request (JSON)", func(t *testing.T) {
		requestBody := `{"ping": "pong"}`
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		require.NoError(t, err, "Failed to write to gzip writer")
		err = zb.Close()
		require.NoError(t, err, "Failed to close gzip writer")

		w := httptest.NewRecorder()

		req, err := http.NewRequest("POST", "/test-json", buf)
		require.NoError(t, err, "Failed to create request")

		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code, "Status code should be 200")

		expectedResponse := `{"received_body":"{\"ping\": \"pong\"}"}`
		require.JSONEq(t, expectedResponse, w.Body.String(), "Handler should receive decompressed body")
	})

	t.Run("Client accepts gzipped response (JSON)", func(t *testing.T) {
		w := httptest.NewRecorder()

		req, err := http.NewRequest("POST", "/test-json", bytes.NewBufferString(`{"ping": "pong"}`))
		require.NoError(t, err)

		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "gzip", w.Header().Get("Content-Encoding"), "Response should be gzipped")

		zr, err := gzip.NewReader(w.Body)
		require.NoError(t, err, "Failed to create gzip reader")
		defer zr.Close()

		body, err := io.ReadAll(zr)
		require.NoError(t, err, "Failed to read gzipped body")

		expectedBody := `{"received_body":"{\"ping\": \"pong\"}"}`
		require.JSONEq(t, expectedBody, string(body), "Body content mismatch after decompression")
	})

	t.Run("Client accepts gzipped response (HTML)", func(t *testing.T) {
		w := httptest.NewRecorder()

		req, err := http.NewRequest("GET", "/test-html", nil)
		require.NoError(t, err)

		req.Header.Set("Accept-Encoding", "gzip")

		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "gzip", w.Header().Get("Content-Encoding"), "Response should be gzipped")
		require.Contains(t, w.Header().Get("Content-Type"), "text/html", "Content-Type should be HTML")

		zr, err := gzip.NewReader(w.Body)
		require.NoError(t, err, "Failed to create gzip reader")
		defer zr.Close()

		body, err := io.ReadAll(zr)
		require.NoError(t, err, "Failed to read gzipped body")

		require.Equal(t, "<h1>Hello</h1>", string(body), "HTML content mismatch")
	})

	t.Run("Server should not compress response (Plain Text)", func(t *testing.T) {
		w := httptest.NewRecorder()

		req, err := http.NewRequest("GET", "/test-text", nil)
		require.NoError(t, err)

		req.Header.Set("Accept-Encoding", "gzip")

		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		require.Empty(t, w.Header().Get("Content-Encoding"), "Response should NOT be gzipped")

		require.Equal(t, "No compression", w.Body.String(), "Plain text content mismatch")
	})
}
