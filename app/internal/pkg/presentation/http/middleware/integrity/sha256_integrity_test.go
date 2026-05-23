package integrity

import (
	"net/http"
	"testing"

	pkgintegrity "github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/integrity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSHA256Integrity_emptyKey(t *testing.T) {
	assert.Nil(t, NewSHA256Integrity(""))
}

func TestSHA256Integrity_verify(t *testing.T) {
	s := NewSHA256Integrity("key")
	require.NotNil(t, s)
	assert.True(t, s.Matches(http.Header{}))

	body := []byte("data")
	hash := s.Calculate(body)
	h := http.Header{}
	h.Set(HashSHA256Header, hash)
	assert.NoError(t, s.Verify(body, h))
	assert.Equal(t, HashSHA256Header, s.HeaderName())
	assert.Equal(t, pkgintegrity.HashSHA256Header, HashSHA256Header)
}
