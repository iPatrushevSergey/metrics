package logger

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/logger"
	"github.com/stretchr/testify/assert"
)

func TestLogger_middleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Logger(logger.NewNopLogger(), nil))
	r.POST("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/plain", []byte("ok"))
	})

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("in")))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
}
