package model

// AuditEvent is the on-disk JSON projection of an audit event line.
type AuditEvent struct {
	TS        int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}
