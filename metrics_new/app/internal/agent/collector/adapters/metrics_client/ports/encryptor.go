package ports

// Encryptor interface for data encryption.
type Encryptor interface {
	Seal(plaintext []byte) (payload []byte, contentType string, headers map[string]string, err error)
}
