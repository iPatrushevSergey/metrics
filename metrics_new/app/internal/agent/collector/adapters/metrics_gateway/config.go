package metrics_gateway

import "time"

// MetricsGatewayConfig holds HTTP transport settings for the metrics gateway.
type MetricsGatewayConfig struct {
	Address     string        `mapstructure:"address"`
	HTTPTimeout time.Duration `mapstructure:"http_timeout"`
}
