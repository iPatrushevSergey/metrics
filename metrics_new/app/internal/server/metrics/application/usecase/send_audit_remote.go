package usecase

import (
	"context"
	"fmt"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/port"
)

// SendAuditRemote is a use case for sending audit events to a remote endpoint.
type SendAuditRemote struct {
	gateway port.AuditGateway
}

// NewSendAuditRemote creates a new use case for sending audit events to a remote endpoint.
func NewSendAuditRemote(gateway port.AuditGateway) port.UseCase[dto.AuditEvent, struct{}] {
	return &SendAuditRemote{gateway: gateway}
}

// Execute sends the event to the remote endpoint.
func (uc *SendAuditRemote) Execute(ctx context.Context, event dto.AuditEvent) (struct{}, error) {
	if err := uc.gateway.Send(ctx, event); err != nil {
		return struct{}{}, fmt.Errorf("%w: %v", application.ErrInternal, err)
	}
	return struct{}{}, nil
}
