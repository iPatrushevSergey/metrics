// Package cryption provides request-body decryption middleware with pluggable strategies.
package cryption

import (
	"bytes"
	"io"
	"net/http"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/port"

	"github.com/gin-gonic/gin"
)

// Decryptor decrypts ciphertext bytes.
type Decryptor interface {
	Matches(headers http.Header) bool
	Decrypt(body []byte) ([]byte, error)
}

// DecryptRequests decrypts requests.
func DecryptRequests(log port.Logger, decryptors ...Decryptor) gin.HandlerFunc {
	if len(decryptors) == 0 {
		return func(c *gin.Context) { c.Next() }
	}

	return func(c *gin.Context) {
		var dec Decryptor
		for _, d := range decryptors {
			if d != nil && d.Matches(c.Request.Header) {
				dec = d
				break
			}
		}
		if dec == nil {
			c.Next()
			return
		}

		if c.Request.Body == nil {
			c.Next()
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			if log != nil {
				log.Error("http middleware: read encrypted body", "error", err)
			}
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		plain, err := dec.Decrypt(body)
		if err != nil {
			if log != nil {
				log.Error("DecryptRequests middleware: decrypt body", "error", err)
			}
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewReader(plain))
		c.Request.ContentLength = int64(len(plain))
		c.Next()
	}
}
