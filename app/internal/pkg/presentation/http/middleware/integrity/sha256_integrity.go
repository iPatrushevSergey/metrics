package integrity

import (
	"net/http"
	"strings"

	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/integrity"
)

// HashSHA256Header is a header for SHA256 hash.
const HashSHA256Header = "HashSHA256"

// SHA256Integrity is a implementation of IntegrityHasher for SHA256.
type SHA256Integrity struct {
	key    string
	hasher integrity.SHA256Hasher
}

// NewSHA256Integrity creates a new SHA256Integrity.
func NewSHA256Integrity(key string) *SHA256Integrity {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil
	}
	return &SHA256Integrity{
		key:    key,
		hasher: integrity.NewSHA256Hasher(key),
	}
}

// Matches key configured, headers unused.
func (s *SHA256Integrity) Matches(_ http.Header) bool {
	return s != nil && s.key != ""
}

// Verify verifies the SHA256Integrity.
func (s *SHA256Integrity) Verify(body []byte, headers http.Header) error {
	return s.hasher.VerifyHash(body, headers.Get(HashSHA256Header))
}

// Calculate calculates the SHA256 hash of the body.
func (s *SHA256Integrity) Calculate(body []byte) string {
	return s.hasher.CalculateHash(body)
}

// HeaderName returns the header name for the SHA256Integrity.
func (s *SHA256Integrity) HeaderName() string {
	return HashSHA256Header
}
