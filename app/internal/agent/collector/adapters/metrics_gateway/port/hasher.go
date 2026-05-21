package port

// Hasher interface for data integrity hashing.
type Hasher interface {
	CalculateHash(body []byte) string
	VerifyHash(body []byte, providedHex string) error
}
