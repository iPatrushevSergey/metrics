package metricsgateway

import "time"

// MetricsGatewayConfig holds HTTP transport settings for the metrics gateway.
type MetricsGatewayConfig struct {
	Address     string        `mapstructure:"address"`
	HTTPTimeout time.Duration `mapstructure:"http_timeout"`
	// RealIP is sent as X-Real-IP; auto-detected from host interfaces when empty.
	RealIP string `mapstructure:"real_ip"`
}
