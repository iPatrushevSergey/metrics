package usecase

import (
	"context"
	"fmt"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/port"
)

// CreateAudit is a use case for creating audit events at a remote endpoint.
type CreateAudit struct {
	gateway port.AuditGateway
}

// NewCreateAudit creates a use case for remote audit event creation.
func NewCreateAudit(gateway port.AuditGateway) port.UseCase[dto.AuditEvent, struct{}] {
	return &CreateAudit{gateway: gateway}
}

// Execute creates the audit event at the remote endpoint.
func (uc *CreateAudit) Execute(ctx context.Context, event dto.AuditEvent) (struct{}, error) {
	if err := uc.gateway.CreateAudit(ctx, event); err != nil {
		return struct{}{}, fmt.Errorf("%w: %v", application.ErrInternal, err)
	}
	return struct{}{}, nil
}
