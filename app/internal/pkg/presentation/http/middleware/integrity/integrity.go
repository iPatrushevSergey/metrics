// Package integrity verifies request digests and signs responses using pluggable hashers.
package integrity

import (
	"bytes"
	"io"
	"net/http"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/port"

	"github.com/gin-gonic/gin"
)

// IntegrityHasher is a interface for integrity hashers.
type IntegrityHasher interface {
	Matches(headers http.Header) bool
	Verify(body []byte, headers http.Header) error
	Calculate(body []byte) string
	HeaderName() string
}

// Integrity destination middleware for integrity hashers.
func Integrity(log port.Logger, hashers ...IntegrityHasher) gin.HandlerFunc {
	if len(hashers) == 0 {
		return func(c *gin.Context) { c.Next() }
	}

	return func(c *gin.Context) {
		var hasher IntegrityHasher
		for _, h := range hashers {
			if h != nil && h.Matches(c.Request.Header) {
				hasher = h
				break
			}
		}
		if hasher == nil {
			c.Next()
			return
		}

		if c.Request.Body != nil && c.Request.ContentLength != 0 {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err != nil {
				if log != nil {
					log.Error("integrity middleware: read request body for integrity", "error", err)
				}
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			if len(bodyBytes) > 0 {
				if err := hasher.Verify(bodyBytes, c.Request.Header); err != nil {
					if log != nil {
						log.Error("integrity middleware: integrity verify request", "error", err)
					}
					c.AbortWithStatus(http.StatusBadRequest)
					return
				}
			}
		}

		hashWriter := newHashResponseWriter(c.Writer, hasher)
		c.Writer = hashWriter

		c.Next()

		if !hashWriter.wroteHeader {
			hashWriter.Header().Set(hasher.HeaderName(), hasher.Calculate([]byte{}))
			hashWriter.ResponseWriter.WriteHeader(hashWriter.statusCode)
		}
	}
}

// hashResponseWriter is a writer for hash response.
type hashResponseWriter struct {
	gin.ResponseWriter
	hasher      IntegrityHasher
	body        *bytes.Buffer
	statusCode  int
	wroteHeader bool
}

// newHashResponseWriter creates a new hash response writer.
func newHashResponseWriter(w gin.ResponseWriter, hasher IntegrityHasher) *hashResponseWriter {
	return &hashResponseWriter{
		ResponseWriter: w,
		hasher:         hasher,
		body:           bytes.NewBuffer(nil),
		statusCode:     http.StatusOK,
	}
}

// WriteHeader writes the header to the response.
func (w *hashResponseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}
	w.statusCode = statusCode
}

// Write writes the data to the response.
func (w *hashResponseWriter) Write(p []byte) (int, error) {
	if !w.wroteHeader {
		w.body.Write(p)
		w.Header().Set(w.hasher.HeaderName(), w.hasher.Calculate(w.body.Bytes()))
		w.ResponseWriter.WriteHeader(w.statusCode)
		w.wroteHeader = true
	}
	return w.ResponseWriter.Write(p)
}

// WriteString writes the string to the response.
func (w *hashResponseWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}
