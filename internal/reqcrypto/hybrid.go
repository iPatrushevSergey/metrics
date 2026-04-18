// Package reqcrypto encrypts HTTP metric payloads for agent→server (RSA-OAEP + AES-GCM).
package reqcrypto

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
	// HeaderName marks requests with a hybrid-encrypted body (after gzip: ciphertext).
	HeaderName  = "X-Encrypted"
	HeaderValue = "1"
)

const (
	aesKeySize     = 32
	gcmNonceSize   = 12 // standard AES-GCM IV length in Go's cipher.NewGCM
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

// Open decrypts data produced by Seal.
func Open(priv *rsa.PrivateKey, data []byte) ([]byte, error) {
	if priv == nil {
		return nil, errors.New("nil private key")
	}
	if len(data) < 4 {
		return nil, errors.New("truncated crypto payload")
	}
	ekLen := int(binary.BigEndian.Uint32(data[:4]))
	if ekLen <= 0 || len(data) < 4+ekLen {
		return nil, errors.New("invalid encrypted key length")
	}
	ek := data[4 : 4+ekLen]
	rest := data[4+ekLen:]
	if len(rest) < gcmNonceSize {
		return nil, errors.New("truncated nonce")
	}
	nonce := rest[:gcmNonceSize]
	sealed := rest[gcmNonceSize:]

	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, priv, ek, nil)
	if err != nil {
		return nil, fmt.Errorf("rsa decrypt aes key: %w", err)
	}
	if len(aesKey) != aesKeySize {
		return nil, errors.New("unexpected aes key size")
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	plain, err := gcm.Open(nil, nonce, sealed, nil)
	if err != nil {
		return nil, fmt.Errorf("aes-gcm open: %w", err)
	}
	return plain, nil
}
