package compression

import (
	"compress/gzip"
	"fmt"
	"io"
)

// GzipCompressor implements gzip compression.
type GzipCompressor struct{}

// NewGzipCompressor initializes GzipCompressor.
func NewGzipCompressor() GzipCompressor {
	return GzipCompressor{}
}

// ContentEncoding returns the content encoding.
func (GzipCompressor) ContentEncoding() string {
	return "gzip"
}

// DecompressReader decompresses the reader.
func (GzipCompressor) DecompressReader(r io.Reader) (io.ReadCloser, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("gzip decompress: %w", err)
	}
	return zr, nil
}

// CompressWriter compresses the writer.
func (GzipCompressor) CompressWriter(w io.Writer) io.WriteCloser {
	return gzip.NewWriter(w)
}
