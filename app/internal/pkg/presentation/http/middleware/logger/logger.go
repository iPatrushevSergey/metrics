package logger

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/port"

	"github.com/gin-gonic/gin"
)

// LogFormatter abstraction for logging HTTP request details.
type LogFormatter interface {
	Log(log port.Logger, params LogParams)
}

// LogParams contains data available for logging.
type LogParams struct {
	Ctx          *gin.Context
	Duration     time.Duration
	RequestBody  []byte
	ResponseBody *bytes.Buffer
}

// Logger middleware with injected formatter.
func Logger(log port.Logger, formatter LogFormatter) gin.HandlerFunc {
	if formatter == nil {
		formatter = &DefaultLogFormatter{}
	}

	return func(c *gin.Context) {
		start := time.Now()

		var requestBody []byte
		if c.Request.Body != nil && c.Request.ContentLength != 0 {
			var readErr error
			requestBody, readErr = io.ReadAll(c.Request.Body)
			if readErr != nil {
				log.Error("failed to read request body", "error", readErr)
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewReader(requestBody))
		}

		w := &responseBodyWriter{body: bytes.NewBuffer(nil), ResponseWriter: c.Writer}
		c.Writer = w

		c.Next()

		formatter.Log(log, LogParams{
			Ctx:          c,
			Duration:     time.Since(start),
			RequestBody:  requestBody,
			ResponseBody: w.body,
		})
	}
}

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func (r responseBodyWriter) WriteString(s string) (int, error) {
	r.body.WriteString(s)
	return r.ResponseWriter.WriteString(s)
}
