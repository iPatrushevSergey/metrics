package audit_gateway

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/retry"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/port"
)

// preparedRequest contains the URL, body bytes, and headers for sending the request.
type preparedRequest struct {
	url     string
	body    []byte
	headers http.Header
}

// AuditRemoteGateway sends audit events via HTTP POST.
type AuditRemoteGateway struct {
	httpClient  *http.Client
	endpointURL string
	retryOpts   []retry.RetryOption
}

// NewAuditRemoteGateway creates a gateway for the configured audit HTTP endpoint.
func NewAuditRemoteGateway(cfg AuditGatewayConfig, httpClient *http.Client, retryOpts ...retry.RetryOption) (*AuditRemoteGateway, error) {
	if httpClient == nil {
		return nil, fmt.Errorf("http client is required")
	}
	parsedURL, err := url.ParseRequestURI(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid audit URL: %w", err)
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, fmt.Errorf("invalid audit URL: scheme and host are required")
	}
	return &AuditRemoteGateway{
		httpClient:  httpClient,
		endpointURL: parsedURL.String(),
		retryOpts:   retryOpts,
	}, nil
}

// CreateAudit posts the audit event as JSON.
func (g *AuditRemoteGateway) CreateAudit(ctx context.Context, e dto.AuditEvent) error {
	prepared, err := g.createAuditPrepare(e)
	if err != nil {
		return err
	}

	var resp *http.Response

	err = retry.DoWithRetry(ctx, func() error {
		r, sendErr := g.createAuditSend(ctx, prepared)
		if sendErr != nil {
			return sendErr
		}
		resp = r
		return nil
	}, g.retryOpts...)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return g.createAuditProcessResponse(resp)
}

var _ port.AuditGateway = (*AuditRemoteGateway)(nil)
