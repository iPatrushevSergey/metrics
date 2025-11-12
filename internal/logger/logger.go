package logger

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var Log *zap.Logger = zap.NewNop()

func Initialize(level string) (*zap.Logger, error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	Log = zl
	return zl, nil
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

func ZapLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		var requestBodyBytes []byte

		if ctx.Request.Body != nil && ctx.Request.ContentLength > 0 {
			requestBodyBytes, _ = io.ReadAll(ctx.Request.Body)
			ctx.Request.Body = io.NopCloser(bytes.NewBuffer(requestBodyBytes))
		}

		w := &responseBodyWriter{body: bytes.NewBufferString(""), ResponseWriter: ctx.Writer}
		ctx.Writer = w

		ctx.Next()

		duration := time.Since(start)

		Log.Info(
			"HTTP Request processed",
			zap.String("URI", ctx.Request.RequestURI),
			zap.String("method", ctx.Request.Method),
			zap.Duration("duration", duration),
			zap.Int("status", ctx.Writer.Status()),
			zap.Int("size", ctx.Writer.Size()),
			zap.String("request_body", string(requestBodyBytes)),
			zap.String("response_body", w.body.String()),
		)
	}
}
