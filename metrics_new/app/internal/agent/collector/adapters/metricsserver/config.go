package metricsserver

import "time"

// MetricsClientConfig holds HTTP transport settings for MetricsClient.
type MetricsClientConfig struct {
	Address     string        `mapstructure:"address"`
	HTTPTimeout time.Duration `mapstructure:"http_timeout"`
}
