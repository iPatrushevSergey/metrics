package port

import (
	"context"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/dto"
)

// AuditGateway sends audit events to a remote HTTP endpoint.
type AuditGateway interface {
	Send(ctx context.Context, e dto.AuditEvent) error
}
