package compression

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldCompress(t *testing.T) {
	assert.True(t, shouldCompress("application/json; charset=utf-8"))
	assert.True(t, shouldCompress("text/html"))
	assert.False(t, shouldCompress("image/png"))
	assert.False(t, shouldCompress(""))
}

func TestCompress_responseGzip(t *testing.T) {
	gin.SetMode(gin.TestMode)
	comp, err := NewGzipCompressor(gzip.DefaultCompression)
	require.NoError(t, err)

	r := gin.New()
	r.Use(Compress(logger.NewNopLogger(), comp))
	r.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		_, _ = c.Writer.Write([]byte(`{"ok":true}`))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))

	gr, err := gzip.NewReader(w.Body)
	require.NoError(t, err)
	defer gr.Close()
	out, err := io.ReadAll(gr)
	require.NoError(t, err)
	assert.JSONEq(t, `{"ok":true}`, string(out))
}

func TestCompress_requestGzip(t *testing.T) {
	gin.SetMode(gin.TestMode)
	comp, err := NewGzipCompressor(gzip.DefaultCompression)
	require.NoError(t, err)

	var got string
	r := gin.New()
	r.Use(Compress(logger.NewNopLogger(), comp))
	r.POST("/", func(c *gin.Context) {
		b, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		got = string(b)
		c.Status(http.StatusOK)
	})

	var gz bytes.Buffer
	zw := gzip.NewWriter(&gz)
	_, _ = zw.Write([]byte(`{"in":1}`))
	_ = zw.Close()

	req := httptest.NewRequest(http.MethodPost, "/", &gz)
	req.Header.Set("Content-Encoding", "gzip")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"in":1}`, got)
}
