package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

/////////// Writer /////////////

type compressWriter struct {
	gin.ResponseWriter
	zw *gzip.Writer

	statusCode     int
	wroteHeader    bool
	shouldCompress bool
}

func newCompressWriter(w gin.ResponseWriter) *compressWriter {
	return &compressWriter{
		ResponseWriter: w,
		zw:             gzip.NewWriter(w),
		statusCode:     http.StatusOK,
	}
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if c.wroteHeader {
		return
	}
	c.statusCode = statusCode
}

func (c *compressWriter) Write(p []byte) (int, error) {
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

func (c *compressWriter) WriteString(s string) (int, error) {
	return c.Write([]byte(s))
}

func (c *compressWriter) Close() error {
	if c.shouldCompress {
		return c.zw.Close()
	}
	return nil
}

///////// Middleware ///////////

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

		cw := newCompressWriter(c.Writer)
		defer cw.Close()

		originalWriter := c.Writer
		c.Writer = cw

		defer func() {
			c.Writer = originalWriter
		}()

		c.Next()
	}
}
