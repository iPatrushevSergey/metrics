package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iPatrushevSergey/metrics/internal/hash"
)

// HashMiddleware creates middleware for hash verification and response signing
func HashMiddleware(key string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip hash verification if no key is provided
		if key == "" {
			c.Next()
			return
		}

		// Read request body for hash verification
		// Only verify hash for requests with body (POST, PUT, etc.)
		if c.Request.Body != nil && c.Request.ContentLength != 0 {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}

			// Restore body for handlers
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			// Verify hash from request header (only if body is not empty)
			if len(bodyBytes) > 0 {
				providedHash := c.Request.Header.Get("HashSHA256")
				if err := hash.VerifyHash(bodyBytes, key, providedHash); err != nil {
					c.AbortWithStatus(http.StatusBadRequest)
					return
				}
			}
		}

		// Create response writer to capture response body
		hw := newHashResponseWriter(c.Writer, key)

		c.Writer = hw

		c.Next()

		if !hw.wroteHeader {
			// Calculate hash from empty body
			responseHash := hash.CalculateHash([]byte{}, key)
			hw.Header().Set("HashSHA256", responseHash)
			hw.ResponseWriter.WriteHeader(hw.statusCode)
		}
	}
}

// hashResponseWriter captures response body for hash calculation
type hashResponseWriter struct {
	gin.ResponseWriter
	body        *bytes.Buffer
	key         string
	statusCode  int
	wroteHeader bool
}

// newHashResponseWriter new writer for hash calculation
func newHashResponseWriter(w gin.ResponseWriter, key string) *hashResponseWriter {
	return &hashResponseWriter{
		ResponseWriter: w,
		body:           &bytes.Buffer{},
		key:            key,
		statusCode:     http.StatusOK,
	}
}

// WriteHeader write header for hash calculation
func (w *hashResponseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}
	w.statusCode = statusCode
}

// Write write data for hash calculation
func (w *hashResponseWriter) Write(p []byte) (int, error) {
	if !w.wroteHeader {
		w.body.Write(p)
		responseHash := hash.CalculateHash(w.body.Bytes(), w.key)
		w.Header().Set("HashSHA256", responseHash)
		w.ResponseWriter.WriteHeader(w.statusCode)
		w.wroteHeader = true
	} else {
		w.body.Write(p)
	}
	return w.ResponseWriter.Write(p)
}

// WriteString write string for hash calculation
func (w *hashResponseWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}
