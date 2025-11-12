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

func ZapLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		var bodyBytes []byte

		if ctx.Request.Body != nil && ctx.Request.ContentLength > 0 {
			bodyBytes, _ = io.ReadAll(ctx.Request.Body)
			ctx.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		ctx.Next()

		duration := time.Since(start)

		Log.Info(
			"HTTP Request processed",
			zap.String("URI", ctx.Request.RequestURI),
			zap.String("method", ctx.Request.Method),
			zap.Duration("duration", duration),
			zap.Int("status", ctx.Writer.Status()),
			zap.Int("size", ctx.Writer.Size()),
		)

		if len(bodyBytes) > 0 {
			Log.Info(
				"Request body logged",
				zap.String("request_body", string(bodyBytes)),
			)
		}
	}
}
