// Package middleware provides Gin middleware: gzip, hash verification, request logging.
package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/iPatrushevSergey/metrics/internal/logger"
)

var (
	gzipReaderPool sync.Pool
	gzipWriterPool sync.Pool
)

// GzipGinMiddleware middleware for gzip compression
func GzipGinMiddleware(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Incoming Request Processing (Decompress)
		contentEncoding := c.Request.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := getGzipReader(c.Request.Body)
			if err != nil {
				log.Error(
					"Failed to create gzip reader",
					zap.Error(err),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
				)
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			c.Request.Body = cr
			defer putGzipReader(cr)
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

func getGzipReader(r io.Reader) (*gzip.Reader, error) {
	if v := gzipReaderPool.Get(); v != nil {
		zr := v.(*gzip.Reader)
		if err := zr.Reset(r); err != nil {
			gzipReaderPool.Put(v)
			return nil, err
		}
		return zr, nil
	}
	return gzip.NewReader(r)
}

func putGzipReader(zr *gzip.Reader) {
	_ = zr.Close()
	gzipReaderPool.Put(zr)
}

// newCompressResponseWriter new writer for gzip compression (reuses gzip.Writer from pool)
func newCompressResponseWriter(w gin.ResponseWriter) *compressResponseWriter {
	var zw *gzip.Writer
	if v := gzipWriterPool.Get(); v != nil {
		zw = v.(*gzip.Writer)
		zw.Reset(w)
	} else {
		zw = gzip.NewWriter(w)
	}
	return &compressResponseWriter{
		ResponseWriter: w,
		zw:             zw,
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

// Close close writer for gzip compression and return to pool
func (c *compressResponseWriter) Close() error {
	if c.shouldCompress {
		err := c.zw.Close()
		c.zw.Reset(nil)
		gzipWriterPool.Put(c.zw)
		return err
	}
	c.zw.Reset(nil)
	gzipWriterPool.Put(c.zw)
	return nil
}
