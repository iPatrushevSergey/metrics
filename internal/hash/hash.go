package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// CalculateHash calculates SHA256 hash from body + key
// Formula: hash(body + key)
func CalculateHash(body []byte, key string) string {
	if key == "" {
		return ""
	}

	data := append(body, []byte(key)...)

	hash := sha256.Sum256(data)

	return hex.EncodeToString(hash[:])
}

// VerifyHash verifies that the calculated hash matches the provided hash
func VerifyHash(body []byte, key string, providedHash string) error {
	// If no key is configured, hash verification is not required
	if key == "" {
		return nil
	}

	// If key is set, hash must be provided
	if providedHash == "" {
		return fmt.Errorf("hash is required when key is configured")
	}

	calculatedHash := CalculateHash(body, key)
	if calculatedHash != providedHash {
		return fmt.Errorf("hash mismatch: expected %s, got %s", calculatedHash, providedHash)
	}
	return nil
}
