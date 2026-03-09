package hash

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateHash_emptyKey(t *testing.T) {
	got := CalculateHash([]byte("body"), "")
	assert.Empty(t, got)
}

func TestCalculateHash_withKey(t *testing.T) {
	got := CalculateHash([]byte("data"), "secret")
	assert.Len(t, got, 64)
	assert.NotEmpty(t, got)
	got2 := CalculateHash([]byte("data"), "secret")
	assert.Equal(t, got, got2, "CalculateHash should be deterministic")
}

func TestVerifyHash_emptyKey(t *testing.T) {
	err := VerifyHash([]byte("x"), "", "any")
	require.NoError(t, err)
}

func TestVerifyHash_keySet_noHash(t *testing.T) {
	err := VerifyHash([]byte("x"), "key", "")
	require.Error(t, err)
}

func TestVerifyHash_match(t *testing.T) {
	body := []byte("test")
	key := "k"
	h := CalculateHash(body, key)
	err := VerifyHash(body, key, h)
	require.NoError(t, err)
}

func TestVerifyHash_mismatch(t *testing.T) {
	err := VerifyHash([]byte("a"), "k", "wrong")
	require.Error(t, err)
}
