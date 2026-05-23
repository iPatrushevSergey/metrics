package integrity

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type nopLogger struct{}

func (nopLogger) Debug(string, ...any) {}
func (nopLogger) Info(string, ...any)  {}
func (nopLogger) Warn(string, ...any)  {}
func (nopLogger) Error(string, ...any) {}
func (nopLogger) Sync() error          { return nil }

func TestIntegrity_withHasher(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSHA256Integrity("secret")
	require.NotNil(t, h)

	body := []byte(`{"id":"x"}`)
	hash := h.Calculate(body)

	r := gin.New()
	r.Use(Integrity(nopLogger{}, h))
	r.POST("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json", []byte(`{"ok":true}`))
	})

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set(HashSHA256Header, hash)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get(HashSHA256Header))
}

func TestIntegrity_noHashers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Integrity(nil))
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
