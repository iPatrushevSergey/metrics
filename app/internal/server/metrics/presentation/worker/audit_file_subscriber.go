package worker

import (
	"context"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/port"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/presentation/factory"
)

// AuditFileSubscriber is a worker for recording audit events to a file.
type AuditFileSubscriber struct {
	events   <-chan dto.AuditEvent
	useCases factory.UseCaseFactory
	log      port.Logger
}

// NewAuditFileSubscriber creates a new worker for recording audit events to a file.
func NewAuditFileSubscriber(
	events <-chan dto.AuditEvent,
	useCases factory.UseCaseFactory,
	log port.Logger,
) *AuditFileSubscriber {
	return &AuditFileSubscriber{events: events, useCases: useCases, log: log}
}

// Run records audit events to the file.
func (w *AuditFileSubscriber) Run(ctx context.Context) {
	uc := w.useCases.RecordAuditToFileUseCase()
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
				w.log.Error("audit file delivery failed", "error", err)
			}
		}
	}
}
