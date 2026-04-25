// Package payloadseal implements RSA-OAEP + AES-GCM sealing for agent->server payloads.
package payloadseal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	// HeaderName marks requests with a hybrid-encrypted body.
	HeaderName  = "X-Encrypted"
	HeaderValue = "1"
)

const (
	aesKeySize   = 32
	gcmNonceSize = 12
)

// Seal encrypts plaintext with a random AES-256-GCM key and wraps the key with RSA-OAEP (SHA-256).
// Wire: u32be(len(ek)) || ek || nonce || ciphertext+tag.
func Seal(pub *rsa.PublicKey, plaintext []byte) ([]byte, error) {
	if pub == nil {
		return nil, errors.New("nil public key")
	}
	aesKey := make([]byte, aesKeySize)
	if _, err := rand.Read(aesKey); err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcmNonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	sealed := gcm.Seal(nil, nonce, plaintext, nil)

	ek, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pub, aesKey, nil)
	if err != nil {
		return nil, fmt.Errorf("rsa encrypt aes key: %w", err)
	}
	if len(ek) > 0xffffff {
		return nil, errors.New("encrypted key too large")
	}
	out := make([]byte, 4+len(ek)+len(nonce)+len(sealed))
	binary.BigEndian.PutUint32(out[:4], uint32(len(ek)))
	copy(out[4:], ek)
	copy(out[4+len(ek):], nonce)
	copy(out[4+len(ek)+len(nonce):], sealed)
	return out, nil
}
