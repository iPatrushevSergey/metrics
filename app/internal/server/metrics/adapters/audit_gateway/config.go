package auditgateway

import "time"

// AuditGatewayConfig holds HTTP transport settings for the audit remote gateway.
type AuditGatewayConfig struct {
	URL         string        `mapstructure:"url"`
	HTTPTimeout time.Duration `mapstructure:"http_timeout"`
}
