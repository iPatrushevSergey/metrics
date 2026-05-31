package dto

// AuditEventRequest is the HTTP request body for audit event payload.
type AuditEventRequest struct {
	TS        int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}
