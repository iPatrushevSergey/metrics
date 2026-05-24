package metrics_gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	httpdto "github.com/iPatrushevSergey/metrics/app/internal/agent/collector/adapters/metrics_gateway/dto"
	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/application/dto"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/encryption"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/http_client"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/integrity"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/netutil"
)

// metricsUpdateBatchPrepare prepares a request for the metrics update batch.
func (c *metricsGateway) metricsUpdateBatchPrepare(metrics []dto.MetricUpdateInput) (preparedRequest, error) {
	var prepReq preparedRequest
	requestPayload := make([]httpdto.MetricUpdateRequest, 0, len(metrics))
	for _, m := range metrics {
		requestPayload = append(requestPayload, httpdto.MetricUpdateRequest{
			ID:    m.ID,
			MType: m.MType,
			Delta: m.Delta,
			Value: m.Value,
		})
	}

	jsonPlaintext, err := json.Marshal(requestPayload)
	if err != nil {
		return prepReq, fmt.Errorf("marshal batch: %w", err)
	}
	if len(jsonPlaintext) == 0 {
		return prepReq, fmt.Errorf("marshal batch produced empty JSON")
	}

	fullURL := c.baseURL + "/updates"

	protected, sealed, err := c.encryptor.Seal(jsonPlaintext)
	if err != nil {
		return prepReq, fmt.Errorf("seal payload: %w", err)
	}

	headers := make(http.Header)
	if sealed {
		headers.Set("Content-Type", "application/octet-stream")
		headers.Set(encryption.XEncryptedHeader, encryption.XEncryptedValue)
	} else {
		headers.Set("Content-Type", "application/json")
	}

	var buf bytes.Buffer
	zw := c.compressor.CompressWriter(&buf)
	if _, err := zw.Write(protected); err != nil {
		return prepReq, fmt.Errorf("compress write: %w", err)
	}
	if err := zw.Close(); err != nil {
		return prepReq, fmt.Errorf("compress close: %w", err)
	}
	compressedBody := buf.Bytes()

	headers.Set("Content-Encoding", c.compressor.ContentEncoding())
	headers.Set("Accept-Encoding", c.compressor.ContentEncoding())
	if c.realIP != "" {
		headers.Set(netutil.RealIPHeader, c.realIP)
	}
	if hash := c.hasher.CalculateHash(jsonPlaintext); hash != "" {
		headers.Set(integrity.HashSHA256Header, hash)
	}

	return preparedRequest{
		url:     fullURL,
		body:    compressedBody,
		headers: headers,
	}, nil
}

// metricsUpdateBatchSend sends a request for the metrics update batch.
func (c *metricsGateway) metricsUpdateBatchSend(ctx context.Context, prepReq preparedRequest) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, prepReq.url, bytes.NewReader(prepReq.body))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header = prepReq.headers.Clone()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusOK {
		return resp, nil
	}

	defer resp.Body.Close()
	body, readErr := io.ReadAll(resp.Body)
	msg := string(body)
	if readErr != nil {
		msg = readErr.Error()
	}

	if http_client.IsRetriableHTTPStatus(resp.StatusCode) {
		return nil, &http_client.RetriableStatusError{Code: resp.StatusCode}
	}
	return nil, fmt.Errorf("HTTP %s: %s", resp.Status, msg)
}

// metricsUpdateBatchProcessResponse processes the response for the metrics update batch.
func (c *metricsGateway) metricsUpdateBatchProcessResponse(resp *http.Response) error {
	var (
		bodyReader io.ReadCloser
		err        error
	)
	if resp.Header.Get("Content-Encoding") != c.compressor.ContentEncoding() {
		bodyReader = resp.Body
	} else {
		bodyReader, err = c.compressor.DecompressReader(resp.Body)
		if err != nil {
			return fmt.Errorf("decompress response body: %w", err)
		}
	}
	defer bodyReader.Close()
	_, err = io.Copy(io.Discard, bodyReader)
	return err
}
