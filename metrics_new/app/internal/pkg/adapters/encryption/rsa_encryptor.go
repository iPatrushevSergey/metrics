package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

const (
	// HeaderName marks requests with a hybrid-encrypted body.
	XEncryptedHeader = "X-Encrypted"
	XEncryptedValue  = "1"

	aesKeySize   = 32
	gcmNonceSize = 12 // standard AES-GCM IV length in Go's cipher.NewGCM
)

// RSAEncryptor hybrid-seals with pub and priv keys.
type RSAEncryptor struct {
	pub  *rsa.PublicKey
	priv *rsa.PrivateKey
}

// NewRSAEncryptorWithPublic initializes RSAEncryptor with public key.
func NewRSAEncryptorWithPublic(pub *rsa.PublicKey) RSAEncryptor {
	return RSAEncryptor{pub: pub}
}

// NewRSAEncryptorWithPrivate initializes RSAEncryptor with private key.
func NewRSAEncryptorWithPrivate(priv *rsa.PrivateKey) RSAEncryptor {
	return RSAEncryptor{priv: priv}
}

// LoadRSAPublicKeyFromFile loads an RSA public key from PEM (PKCS#1, PKIX, or X.509 certificate).
func LoadRSAPublicKeyFromFile(path string) (*rsa.PublicKey, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, errors.New("no PEM block in public key file")
	}
	switch block.Type {
	case "RSA PUBLIC KEY":
		return x509.ParsePKCS1PublicKey(block.Bytes)
	case "PUBLIC KEY":
		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		rs, ok := pub.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("expected RSA public key, got %T", pub)
		}
		return rs, nil
	case "CERTIFICATE":
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}
		rs, ok := cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("certificate public key is %T, want RSA", cert.PublicKey)
		}
		return rs, nil
	default:
		return nil, fmt.Errorf("unsupported PEM type %q", block.Type)
	}
}

// LoadRSAPrivateKeyFromFile loads an RSA private key from PEM (PKCS#1 or PKCS#8).
func LoadRSAPrivateKeyFromFile(path string) (*rsa.PrivateKey, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, errors.New("no PEM block in private key file")
	}
	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "PRIVATE KEY":
		k, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		rs, ok := k.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("expected RSA private key, got %T", k)
		}
		return rs, nil
	default:
		return nil, fmt.Errorf("unsupported PEM type %q", block.Type)
	}
}

// Seal encrypts plaintext with a random AES-256-GCM key and wraps the key with RSA-OAEP (SHA-256).
// Wire: u32be(len(ek)) || ek || nonce || ciphertext+tag.
func (e RSAEncryptor) Seal(plaintext []byte) ([]byte, bool, error) {
	if e.pub == nil {
		return plaintext, false, nil
	}
	pub := e.pub

	aesKey := make([]byte, aesKeySize)
	if _, err := rand.Read(aesKey); err != nil {
		return nil, false, err
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, false, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, false, err
	}
	nonce := make([]byte, gcmNonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, false, err
	}
	payload := gcm.Seal(nil, nonce, plaintext, nil)

	ek, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pub, aesKey, nil)
	if err != nil {
		return nil, false, fmt.Errorf("rsa encrypt aes key: %w", err)
	}
	if len(ek) > 0xffffff {
		return nil, false, errors.New("encrypted key too large")
	}
	out := make([]byte, 4+len(ek)+len(nonce)+len(payload))
	binary.BigEndian.PutUint32(out[:4], uint32(len(ek)))
	copy(out[4:], ek)
	copy(out[4+len(ek):], nonce)
	copy(out[4+len(ek)+len(nonce):], payload)

	return out, true, nil
}

// Open decrypts data produced by Seal.
func (e RSAEncryptor) Open(data []byte) ([]byte, error) {
	if e.priv == nil {
		return nil, errors.New("no private key")
	}
	priv := e.priv
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
