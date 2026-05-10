package metrics_gateway

import "time"

// GatewayConfig holds HTTP transport settings for the metrics Gateway.
type MetricsGatewayConfig struct {
	Address     string        `mapstructure:"address"`
	HTTPTimeout time.Duration `mapstructure:"http_timeout"`
}
