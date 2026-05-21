package dto

// AuditEvent is a DTO for audit events.
type AuditEvent struct {
	TS        int64
	Metrics   []string
	IPAddress string
}
