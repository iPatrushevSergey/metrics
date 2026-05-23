package encryption

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRSAEncryptor_roundTrip(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	enc := NewRSAEncryptorWithPublic(&key.PublicKey)
	dec := NewRSAEncryptorWithPrivate(key)

	sealed, ok, err := enc.Seal([]byte("hello"))
	require.NoError(t, err)
	require.True(t, ok)

	plain, err := dec.Open(sealed)
	require.NoError(t, err)
	assert.Equal(t, "hello", string(plain))
}

func TestRSAEncryptor_noPublicKey(t *testing.T) {
	enc := RSAEncryptor{}
	out, ok, err := enc.Seal([]byte("plain"))
	require.NoError(t, err)
	assert.False(t, ok)
	assert.Equal(t, []byte("plain"), out)
}

func TestRSAEncryptor_Open_invalid(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	dec := NewRSAEncryptorWithPrivate(key)
	_, err = dec.Open([]byte{1, 2, 3})
	assert.Error(t, err)
}
