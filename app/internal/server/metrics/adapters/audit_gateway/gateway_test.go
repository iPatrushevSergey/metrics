package audit_gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAuditRemoteGateway_validation(t *testing.T) {
	_, err := NewAuditRemoteGateway(AuditGatewayConfig{URL: "http://localhost/audit"}, nil)
	assert.Error(t, err)

	_, err = NewAuditRemoteGateway(AuditGatewayConfig{URL: "not-a-url"}, http.DefaultClient)
	assert.Error(t, err)
}

func TestAuditRemoteGateway_CreateAudit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	gw, err := NewAuditRemoteGateway(AuditGatewayConfig{URL: srv.URL}, srv.Client())
	require.NoError(t, err)

	err = gw.CreateAudit(context.Background(), dto.AuditEvent{TS: 1, Metrics: []string{"a"}})
	assert.NoError(t, err)
}

func TestAuditRemoteGateway_CreateAudit_serverError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	gw, err := NewAuditRemoteGateway(AuditGatewayConfig{URL: srv.URL}, srv.Client())
	require.NoError(t, err)

	err = gw.CreateAudit(context.Background(), dto.AuditEvent{TS: 1})
	assert.Error(t, err)
}
