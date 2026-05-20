package usecase

import (
	"context"
	"fmt"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/application/port"
)

// RecordAuditToFile is a use case for recording audit events to a file.
type RecordAuditToFile struct {
	repo port.AuditFileRepository
}

// NewRecordAuditToFile creates a new use case for recording audit events to a file.
func NewRecordAuditToFile(repo port.AuditFileRepository) port.UseCase[dto.AuditEvent, struct{}] {
	return &RecordAuditToFile{repo: repo}
}

// Execute appends the event to the file.
func (uc *RecordAuditToFile) Execute(ctx context.Context, event dto.AuditEvent) (struct{}, error) {
	if err := uc.repo.Append(ctx, event); err != nil {
		return struct{}{}, fmt.Errorf("%w: %v", application.ErrInternal, err)
	}
	return struct{}{}, nil
}
