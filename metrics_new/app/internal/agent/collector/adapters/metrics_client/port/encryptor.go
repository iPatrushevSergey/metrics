package ports

// Encryptor interface for data encryption.
type Encryptor interface {
	Seal(plaintext []byte) (ciphertext []byte, sealed bool, err error)
}
