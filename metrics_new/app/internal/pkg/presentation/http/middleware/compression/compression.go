package compression

import (
	"io"
	"net/http"
	"strings"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/pkg/presentation/http/port"

	"github.com/gin-gonic/gin"
)

// Compressor defines a strategy for compression algorithms (gzip, brotli, etc.).
type Compressor interface {
	// ContentEncoding returns the header value (e.g. "gzip").
	ContentEncoding() string
	// NewReader returns a reader that decompresses data from r.
	NewReader(r io.Reader) (io.ReadCloser, error)
	// NewWriter returns a writer that compresses data to w.
	NewWriter(w io.Writer) io.WriteCloser
}

// Compress middleware supporting multiple compression strategies.
func Compress(log port.Logger, compressors ...Compressor) gin.HandlerFunc {
	encodingToCompressor := make(map[string]Compressor)
	for _, compressor := range compressors {
		encodingToCompressor[compressor.ContentEncoding()] = compressor
	}

	return func(c *gin.Context) {
		c.Header("Vary", "Accept-Encoding")

		// Decompress Request
		reqEncoding := c.GetHeader("Content-Encoding")
		if reqEncoding != "" {
			if reqCompressor, ok := encodingToCompressor[reqEncoding]; ok {
				compressorReader, err := reqCompressor.NewReader(c.Request.Body)
				if err != nil {
					log.Error("failed to create decompress reader", "error", err, "encoding", reqEncoding)
					c.AbortWithStatus(http.StatusBadRequest)
					return
				}
				c.Request.Body = compressorReader
				defer func() {
					if err := compressorReader.Close(); err != nil {
						log.Warn("failed to close decompress reader", "error", err, "encoding", reqEncoding)
					}
				}()
			}
		}

		// Compress Response
		resEncoding := c.GetHeader("Accept-Encoding")
		var resCompressor Compressor
		for _, compressor := range compressors {
			if strings.Contains(resEncoding, compressor.ContentEncoding()) {
				resCompressor = compressor
				break
			}
		}

		if resCompressor == nil {
			c.Next()
			return
		}

		originalWriter := c.Writer
		writerWithCompressor := &WriterWithCompressor{
			ResponseWriter: c.Writer,
			compressor:     resCompressor,
		}
		c.Writer = writerWithCompressor

		defer func() {
			c.Writer = originalWriter
			if writerWithCompressor.compressorWriter != nil {
				if err := writerWithCompressor.compressorWriter.Close(); err != nil {
					log.Warn(
						"failed to close response compressor writer",
						"error", err,
						"encoding", writerWithCompressor.compressor.ContentEncoding(),
					)
				}
			}
		}()

		c.Next()
	}
}

// WriterWithCompressor creates compressorWriter in WriteHeader only when Content-Type allows compression.
type WriterWithCompressor struct {
	gin.ResponseWriter

	compressor       Compressor
	compressorWriter io.WriteCloser

	wroteHeader    bool
	shouldCompress bool
}

func (cw *WriterWithCompressor) Write(data []byte) (int, error) {
	if !cw.wroteHeader {
		cw.WriteHeader(http.StatusOK)
	}
	if cw.shouldCompress && cw.compressorWriter != nil {
		return cw.compressorWriter.Write(data)
	}
	return cw.ResponseWriter.Write(data)
}

func (cw *WriterWithCompressor) WriteHeader(code int) {
	if cw.wroteHeader {
		return
	}
	cw.wroteHeader = true

	contentType := cw.Header().Get("Content-Type")
	if shouldCompress(contentType) {
		cw.shouldCompress = true
		cw.Header().Set("Content-Encoding", cw.compressor.ContentEncoding())
		cw.Header().Del("Content-Length")
		cw.compressorWriter = cw.compressor.NewWriter(cw.ResponseWriter)
	}

	cw.ResponseWriter.WriteHeader(code)
}

func (cw *WriterWithCompressor) WriteString(s string) (int, error) {
	return cw.Write([]byte(s))
}

func shouldCompress(contentType string) bool {
	if strings.TrimSpace(contentType) == "" {
		return false
	}
	return strings.Contains(contentType, "application/json") ||
		strings.Contains(contentType, "text/html") ||
		strings.Contains(contentType, "text/plain") ||
		strings.Contains(contentType, "application/xml")
}
