package ports

import "io"

// Compressor interface for data compression.
type Compressor interface {
	ContentEncoding() string
	NewReader(r io.Reader) (io.ReadCloser, error)
	NewWriter(w io.Writer) io.WriteCloser
}
