package worker

import (
	"context"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/port"
	presport "github.com/iPatrushevSergey/metrics/app/internal/server/metrics/presentation/port"
)

// AuditRemoteSubscriber is a worker for sending audit events to a remote endpoint.
type AuditRemoteSubscriber struct {
	events   <-chan dto.AuditEvent
	useCases presport.UseCaseFactory
	log      port.Logger
}

// NewAuditRemoteSubscriber creates a new worker for sending audit events to a remote endpoint.
func NewAuditRemoteSubscriber(
	events <-chan dto.AuditEvent,
	useCases presport.UseCaseFactory,
	log port.Logger,
) *AuditRemoteSubscriber {
	return &AuditRemoteSubscriber{events: events, useCases: useCases, log: log}
}

// Run sends audit events to the remote endpoint.
func (w *AuditRemoteSubscriber) Run(ctx context.Context) {
	uc := w.useCases.CreateRemoteAuditUseCase()
	if uc == nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-w.events:
			if !ok {
				return
			}
			if _, err := uc.Execute(ctx, event); err != nil {
				w.log.Error("audit remote delivery failed", "error", err)
			}
		}
	}
}
