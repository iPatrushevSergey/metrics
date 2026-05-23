package cryption

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/encryption"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubDecryptor struct{}

func (stubDecryptor) Matches(h http.Header) bool {
	return h.Get("X-Test-Decrypt") == "1"
}

func (stubDecryptor) Decrypt(body []byte) ([]byte, error) {
	return body, nil
}

func TestDecryptRequests_withDecryptor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := []byte(`{"ok":true}`)

	r := gin.New()
	r.Use(DecryptRequests(logger.NewNopLogger(), stubDecryptor{}))
	r.POST("/", func(c *gin.Context) {
		b, err := c.GetRawData()
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		assert.Equal(t, body, b)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("X-Test-Decrypt", "1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDecryptRequests_noDecryptors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(DecryptRequests(logger.NewNopLogger()))
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRSADecryptor_roundTrip(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	enc := encryption.NewRSAEncryptorWithPublic(&key.PublicKey)
	sealed, ok, err := enc.Seal([]byte("secret"))
	require.NoError(t, err)
	require.True(t, ok)

	dec := NewRSADecryptor(key)
	require.True(t, dec.Matches(http.Header{XEncryptedHeader: []string{XEncryptedValue}}))

	plain, err := dec.Decrypt(sealed)
	require.NoError(t, err)
	assert.Equal(t, "secret", string(plain))
}

func TestRSADecryptor_Matches_nilKey(t *testing.T) {
	var d *RSADecryptor
	assert.False(t, d.Matches(http.Header{XEncryptedHeader: []string{XEncryptedValue}}))
}
