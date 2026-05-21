package audit

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/adapters/repository/file/audit/converter"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/port"
)

var errAuditFileClosed = errors.New("audit file repository closed")

// AuditFileRepository persists audit events to a file.
type AuditFileRepository struct {
	file   *os.File
	mu     sync.Mutex
	closed bool
	conv   converter.AuditConverter
}

// NewAuditFileRepository creates a file repository for the given path.
func NewAuditFileRepository(path string) (*AuditFileRepository, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &AuditFileRepository{
		file: file,
		conv: &converter.AuditConverterImpl{},
	}, nil
}

// Append appends the event to the file.
func (r *AuditFileRepository) Append(ctx context.Context, e dto.AuditEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return errAuditFileClosed
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	payload, err := json.Marshal(r.conv.ToModel(e))
	if err != nil {
		return err
	}
	if _, err = r.file.Write(payload); err != nil {
		return err
	}
	_, err = r.file.Write([]byte{'\n'})
	return err
}

// Close closes the file repository.
func (r *AuditFileRepository) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return nil
	}
	r.closed = true
	return r.file.Close()
}

var _ port.AuditFileRepository = (*AuditFileRepository)(nil)
