package audit_gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/http_client"
	httpdto "github.com/iPatrushevSergey/metrics/app/internal/server/metrics/adapters/audit_gateway/dto"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/dto"
)

// createAuditPrepare prepares a request for creating an audit event.
func (g *AuditRemoteGateway) createAuditPrepare(e dto.AuditEvent) (preparedRequest, error) {
	var prepReq preparedRequest

	body, err := json.Marshal(httpdto.AuditEventRequest{
		TS:        e.TS,
		Metrics:   e.Metrics,
		IPAddress: e.IPAddress,
	})
	if err != nil {
		return prepReq, fmt.Errorf("marshal audit event: %w", err)
	}
	if len(body) == 0 {
		return prepReq, fmt.Errorf("marshal audit event produced empty JSON")
	}

	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")

	return preparedRequest{
		url:     g.endpointURL,
		body:    body,
		headers: headers,
	}, nil
}

// createAuditSend sends a request for the audit event creation.
func (g *AuditRemoteGateway) createAuditSend(ctx context.Context, prepReq preparedRequest) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, prepReq.url, bytes.NewReader(prepReq.body))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header = prepReq.headers.Clone()

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
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

// createAuditProcessResponse processes the response for the audit event creation.
func (g *AuditRemoteGateway) createAuditProcessResponse(resp *http.Response) error {
	_, err := io.Copy(io.Discard, resp.Body)
	return err
}
