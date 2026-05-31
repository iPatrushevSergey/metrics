package integrity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSHA256Hasher_emptyKey(t *testing.T) {
	h := NewSHA256Hasher("")
	assert.Empty(t, h.CalculateHash([]byte("body")))
	assert.NoError(t, h.VerifyHash([]byte("body"), ""))
}

func TestSHA256Hasher_withKey(t *testing.T) {
	h := NewSHA256Hasher("secret")
	hash := h.CalculateHash([]byte("body"))
	require.NotEmpty(t, hash)
	assert.NoError(t, h.VerifyHash([]byte("body"), hash))
	assert.Error(t, h.VerifyHash([]byte("body"), "bad"))
}
