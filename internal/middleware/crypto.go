package middleware

import (
	"bytes"
	"crypto/rsa"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/iPatrushevSergey/metrics/internal/logger"
	"github.com/iPatrushevSergey/metrics/internal/reqcrypto"
)

// CryptoDecryptMiddleware decrypts request bodies marked with X-Encrypted.
func CryptoDecryptMiddleware(priv *rsa.PrivateKey, log logger.Logger) gin.HandlerFunc {
	if priv == nil {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		if c.Request.Header.Get(reqcrypto.HeaderName) != reqcrypto.HeaderValue {
			c.Next()
			return
		}
		if c.Request.Body == nil {
			c.Next()
			return
		}
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Error("crypto middleware read body", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		plain, err := reqcrypto.Open(priv, body)
		if err != nil {
			log.Error("crypto middleware decrypt", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(plain))
		c.Request.ContentLength = int64(len(plain))
		c.Next()
	}
}
