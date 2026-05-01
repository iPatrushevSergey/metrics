package ports

// Hasher derives an integrity header from the plaintext body (agent side: CalculateHash semantics from legacy metrics).
type Hasher interface {
	CalculateHash(body []byte) (headerName, headerValue string, ok bool)
}
