package port

import (
	"context"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/dto"
)

// AuditGateway creates audit events at a remote HTTP endpoint.
type AuditGateway interface {
	CreateAudit(ctx context.Context, e dto.AuditEvent) error
}
