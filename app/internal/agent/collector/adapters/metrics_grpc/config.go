package metricsgrpc

// MetricsGRPCGatewayConfig holds gRPC transport settings for the metrics gateway.
type MetricsGRPCGatewayConfig struct {
	Address string `mapstructure:"grpc_address"`
	RealIP  string `mapstructure:"real_ip"`
}
