package ports

import "io"

// Compressor interface for data compression.
type Compressor interface {
	ContentEncoding() string
	DecompressReader(r io.Reader) (io.ReadCloser, error)
	CompressWriter(w io.Writer) io.WriteCloser
}
