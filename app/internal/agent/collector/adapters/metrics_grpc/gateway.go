// Package metrics_grpc sends metric batches to the server over gRPC.
package metrics_grpc

import (
	"context"
	"fmt"
	"strings"

	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/application/dto"
	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/application/port"
	metricspb "github.com/iPatrushevSergey/metrics/app/internal/pkg/grpc/metrics"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/netutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var _ port.MetricsGateway = (*metricsGateway)(nil)

type metricsGateway struct {
	client metricspb.MetricsClient
	conn   *grpc.ClientConn
	realIP string
}

// NewGateway creates a new metrics gRPC gateway.
func NewGateway(cfg MetricsGRPCGatewayConfig) (*metricsGateway, error) {
	addr := strings.TrimSpace(cfg.Address)
	if addr == "" {
		return nil, fmt.Errorf("grpc address is required")
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("grpc dial: %w", err)
	}

	realIP := strings.TrimSpace(cfg.RealIP)
	if realIP == "" {
		if ip, err := netutil.HostIPv4(); err == nil {
			realIP = ip
		}
	}

	return &metricsGateway{
		client: metricspb.NewMetricsClient(conn),
		conn:   conn,
		realIP: realIP,
	}, nil
}

// Close releases the client connection.
func (g *metricsGateway) Close() error {
	if g.conn == nil {
		return nil
	}
	return g.conn.Close()
}

// MetricsUpdateBatch sends metrics via UpdateMetrics RPC.
func (g *metricsGateway) MetricsUpdateBatch(ctx context.Context, metrics []dto.MetricUpdateInput) error {
	if len(metrics) == 0 {
		return nil
	}

	req := &metricspb.UpdateMetricsRequest{Metrics: make([]*metricspb.Metric, 0, len(metrics))}
	for _, m := range metrics {
		metric := &metricspb.Metric{Id: m.ID}
		switch m.MType {
		case "counter":
			metric.Type = metricspb.Metric_COUNTER
			if m.Delta != nil {
				metric.Delta = *m.Delta
			}
		default:
			metric.Type = metricspb.Metric_GAUGE
			if m.Value != nil {
				metric.Value = *m.Value
			}
		}
		req.Metrics = append(req.Metrics, metric)
	}

	if g.realIP != "" {
		ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(netutil.RealIPHeader, g.realIP))
	}

	_, err := g.client.UpdateMetrics(ctx, req)
	return err
}
