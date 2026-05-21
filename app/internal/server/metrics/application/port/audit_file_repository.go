package port

import (
	"context"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/dto"
)

// AuditFileRepository appends audit events to a log file.
type AuditFileRepository interface {
	Append(ctx context.Context, e dto.AuditEvent) error
	Close() error
}
