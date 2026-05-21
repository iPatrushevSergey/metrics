package cryption

import (
	"crypto/rsa"
	"net/http"

	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/encryption"
)

// HeaderName header name for a hybrid-encrypted body.
const (
	XEncryptedHeader = "X-Encrypted"
	XEncryptedValue  = "1"
)

// RSADecryptor implements hybrid RSA-OAEP + AES-GCM.
type RSADecryptor struct {
	priv *rsa.PrivateKey
}

// NewRSADecryptor returns a RSADecryptor.
func NewRSADecryptor(priv *rsa.PrivateKey) *RSADecryptor {
	return &RSADecryptor{priv: priv}
}

// Matches checks if the headers match the expected format.
func (r *RSADecryptor) Matches(headers http.Header) bool {
	if r == nil || r.priv == nil {
		return false
	}
	return headers.Get(XEncryptedHeader) == XEncryptedValue
}

// Decrypt decrypts the body.
func (r *RSADecryptor) Decrypt(body []byte) ([]byte, error) {
	return encryption.NewRSAEncryptorWithPrivate(r.priv).Open(body)
}
