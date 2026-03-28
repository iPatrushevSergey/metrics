package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/iPatrushevSergey/metrics/internal/hash"
	"github.com/stretchr/testify/require"
)

func TestHashMiddleware_emptyKey_passesThrough(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(HashMiddleware(""))
	r.GET("/x", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))
	require.Equal(t, http.StatusOK, w.Code)
}

func TestHashMiddleware_validHash(t *testing.T) {
	const key = "test-key"
	body := []byte(`{"id":"m","type":"gauge"}`)
	sum := hash.CalculateHash(body, key)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(HashMiddleware(key))
	r.POST("/x", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/x", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HashSHA256", sum)
	req.ContentLength = int64(len(body))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NotEmpty(t, w.Header().Get("HashSHA256"))
}
