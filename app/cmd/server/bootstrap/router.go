package bootstrap

import (
	"compress/gzip"
	"crypto/rsa"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/iPatrushevSergey/metrics/app/internal/pkg/presentation/http/middleware/compression"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/presentation/http/middleware/cryption"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/presentation/http/middleware/integrity"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/presentation/http/middleware/logger"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/port"
	metricrouter "github.com/iPatrushevSergey/metrics/app/internal/server/metrics/presentation/http/router"
)

// NewRouter composes global middleware and module routers.
func NewRouter(ucFactory UseCaseFactory, log port.Logger, key string, priv *rsa.PrivateKey) (*gin.Engine, error) {
	r := gin.New()

	r.Use(gin.Recovery())

	gzipCompressor, err := compression.NewGzipCompressor(gzip.DefaultCompression)
	if err != nil {
		return nil, fmt.Errorf("gzip compressor: %w", err)
	}
	r.Use(compression.Compress(log, gzipCompressor))

	if priv != nil {
		r.Use(cryption.DecryptRequests(log, cryption.NewRSADecryptor(priv)))
	}

	if h := integrity.NewSHA256Integrity(key); h != nil {
		r.Use(integrity.Integrity(log, h))
	}

	r.Use(logger.Logger(log, nil))

	metricrouter.RegisterRoutes(r, ucFactory, log)
	return r, nil
}
