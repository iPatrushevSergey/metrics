package compression

import (
	"bytes"
	"compress/gzip"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipCompressor_roundTrip(t *testing.T) {
	c := NewGzipCompressor()

	var buf bytes.Buffer
	w := c.CompressWriter(&buf)
	_, err := w.Write([]byte("hello"))
	require.NoError(t, err)
	require.NoError(t, w.Close())

	r, err := c.DecompressReader(&buf)
	require.NoError(t, err)
	defer r.Close()

	out, err := io.ReadAll(r)
	require.NoError(t, err)
	assert.Equal(t, "hello", string(out))
	assert.Equal(t, "gzip", c.ContentEncoding())
}

func TestGzipCompressor_invalidReader(t *testing.T) {
	c := NewGzipCompressor()
	_, err := c.DecompressReader(bytes.NewReader([]byte("not gzip")))
	assert.Error(t, err)
}

func TestGzipCompressor_decompressGzipBytes(t *testing.T) {
	var gz bytes.Buffer
	zw := gzip.NewWriter(&gz)
	_, _ = zw.Write([]byte("x"))
	_ = zw.Close()

	c := NewGzipCompressor()
	r, err := c.DecompressReader(&gz)
	require.NoError(t, err)
	defer r.Close()
	out, err := io.ReadAll(r)
	require.NoError(t, err)
	assert.Equal(t, "x", string(out))
}
