package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// GzipMiddleware middleware for gzip compression
func GzipGinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Incoming Request Processing (Decompress)
		contentEncoding := c.Request.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			c.Request.Body = cr
			defer cr.Close()
		}

		// Outgoing response processing (Compress)
		acceptEncoding := c.Request.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")

		if !supportsGzip {
			c.Next()
			return
		}

		cw := newCompressResponseWriter(c.Writer)
		defer cw.Close()

		originalWriter := c.Writer
		c.Writer = cw

		defer func() {
			c.Writer = originalWriter
		}()

		c.Next()
	}
}

// compressResponseWriter writer for gzip compression
type compressResponseWriter struct {
	gin.ResponseWriter
	zw *gzip.Writer

	statusCode     int
	wroteHeader    bool
	shouldCompress bool
}

// newCompressResponseWriter new writer for gzip compression
func newCompressResponseWriter(w gin.ResponseWriter) *compressResponseWriter {
	return &compressResponseWriter{
		ResponseWriter: w,
		zw:             gzip.NewWriter(w),
		statusCode:     http.StatusOK,
	}
}

// WriteHeader write header for gzip compression
func (c *compressResponseWriter) WriteHeader(statusCode int) {
	if c.wroteHeader {
		return
	}
	c.statusCode = statusCode
}

// Write write data for gzip compression
func (c *compressResponseWriter) Write(p []byte) (int, error) {
	if !c.wroteHeader {
		contentType := c.Header().Get("Content-Type")

		if strings.HasPrefix(contentType, "application/json") || strings.HasPrefix(contentType, "text/html") {
			c.shouldCompress = true
			c.Header().Set("Content-Encoding", "gzip")
		}
		c.ResponseWriter.WriteHeader(c.statusCode)
		c.wroteHeader = true
	}

	if c.shouldCompress {
		return c.zw.Write(p)
	}

	return c.ResponseWriter.Write(p)
}

// WriteString write string for gzip compression
func (c *compressResponseWriter) WriteString(s string) (int, error) {
	return c.Write([]byte(s))
}

// Close close writer for gzip compression
func (c *compressResponseWriter) Close() error {
	if c.shouldCompress {
		return c.zw.Close()
	}
	return nil
}
