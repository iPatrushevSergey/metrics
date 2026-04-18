package reqcrypto

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSeal_Open_roundTrip(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	pub := &priv.PublicKey
	plain := []byte(`{"id":"x","type":"gauge","value":1.5}`)

	cipher, err := Seal(pub, plain)
	require.NoError(t, err)
	require.NotEmpty(t, cipher)

	got, err := Open(priv, cipher)
	require.NoError(t, err)
	require.Equal(t, plain, got)
}

func TestOpen_wrongPrivateKey(t *testing.T) {
	privAlice, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	privBob, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	plain := []byte(`{"id":"x"}`)
	cipher, err := Seal(&privAlice.PublicKey, plain)
	require.NoError(t, err)

	_, err = Open(privBob, cipher)
	require.Error(t, err)
}
