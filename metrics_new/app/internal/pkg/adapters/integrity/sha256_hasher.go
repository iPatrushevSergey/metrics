package integrity

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const (
	HashSHA256Header = "HashSHA256"
)

// SHA256Hasher calculates and verifies SHA256 hashes for signed requests.
type SHA256Hasher struct {
	key string
}

// NewSHA256Hasher initializes SHA256Hasher with a signing key.
func NewSHA256Hasher(key string) SHA256Hasher {
	return SHA256Hasher{key: key}
}

// CalculateHash calculates SHA256 hash of the body and the key.
func (h SHA256Hasher) CalculateHash(body []byte) string {
	if h.key == "" {
		return ""
	}
	data := append(body, h.key...)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// VerifyHash verifies the hash of the body and the key.
func (h SHA256Hasher) VerifyHash(body []byte, providedHash string) error {
	if h.key == "" {
		return nil
	}
	if providedHash == "" {
		return fmt.Errorf("hash is required when key is configured")
	}
	calculated := h.CalculateHash(body)
	if calculated != providedHash {
		return fmt.Errorf("hash mismatch: expected %s, got %s", calculated, providedHash)
	}
	return nil
}
