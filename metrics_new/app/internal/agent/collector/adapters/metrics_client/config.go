package metrics_client

import "time"

// MetricsClientConfig holds HTTP transport settings for the metrics server client.
type MetricsClientConfig struct {
	Address     string        `mapstructure:"address"`
	HTTPTimeout time.Duration `mapstructure:"http_timeout"`
}
