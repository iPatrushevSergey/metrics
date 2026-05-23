package encryption

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
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

func TestLoadRSAKeysFromFile(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	dir := t.TempDir()
	privPath := filepath.Join(dir, "priv.pem")
	pubPath := filepath.Join(dir, "pub.pem")

	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	require.NoError(t, os.WriteFile(privPath, privPEM, 0600))

	pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	require.NoError(t, err)
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
	require.NoError(t, os.WriteFile(pubPath, pubPEM, 0600))

	priv, err := LoadRSAPrivateKeyFromFile(privPath)
	require.NoError(t, err)
	require.NotNil(t, priv)

	pub, err := LoadRSAPublicKeyFromFile(pubPath)
	require.NoError(t, err)
	require.NotNil(t, pub)
}

func TestLoadRSAKeysFromFile_invalid(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.pem")
	require.NoError(t, os.WriteFile(bad, []byte("not a key"), 0600))

	_, err := LoadRSAPrivateKeyFromFile(bad)
	assert.Error(t, err)

	_, err = LoadRSAPublicKeyFromFile(bad)
	assert.Error(t, err)

	_, err = LoadRSAPrivateKeyFromFile(filepath.Join(dir, "missing.pem"))
	assert.Error(t, err)
}
