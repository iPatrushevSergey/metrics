package converter

import (
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/adapters/repository/file/audit/model"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/dto"
)

//go:generate goverter gen .

// goverter:converter
// goverter:output:file audit_gen.go
// goverter:output:package converter
type AuditConverter interface {
	ToModel(source dto.AuditEvent) model.AuditEvent
}
