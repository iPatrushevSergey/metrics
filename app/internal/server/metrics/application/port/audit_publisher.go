package port

import (
	"context"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/dto"
)

// AuditPublisher fans out audit events to subscribers.
type AuditPublisher interface {
	Publish(e dto.AuditEvent)
	Subscribe(subscriberID string) (<-chan dto.AuditEvent, error)
	Unsubscribe(subscriberID string)
	Close(ctx context.Context) error
}
