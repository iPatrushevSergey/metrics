package ports

// Encryptor interface for data encryption.
type Encryptor interface {
	Encrypt(plaintext []byte) (payload []byte, contentType string, headers map[string]string, err error)
}
