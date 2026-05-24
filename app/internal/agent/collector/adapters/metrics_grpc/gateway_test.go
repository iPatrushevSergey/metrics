package metrics_grpc

import (
	"context"
	"net"
	"testing"

	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/application/dto"
	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/application/port"
	metricspb "github.com/iPatrushevSergey/metrics/app/internal/pkg/grpc/metrics"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/netutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type stubMetricsServer struct {
	metricspb.UnimplementedMetricsServer
	gotIP   string
	gotSize int
}

func (s *stubMetricsServer) UpdateMetrics(ctx context.Context, req *metricspb.UpdateMetricsRequest) (*metricspb.UpdateMetricsResponse, error) {
	s.gotSize = len(req.GetMetrics())
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get(netutil.RealIPHeader); len(vals) > 0 {
			s.gotIP = vals[0]
		}
	}
	return &metricspb.UpdateMetricsResponse{}, nil
}

func TestMetricsGateway_MetricsUpdateBatch(t *testing.T) {
	wantIP := "192.168.5.10"

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	server := grpc.NewServer()
	stub := &stubMetricsServer{}
	metricspb.RegisterMetricsServer(server, stub)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)

	conn, err := grpc.NewClient(listener.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	gateway := &metricsGateway{
		client: metricspb.NewMetricsClient(conn),
		realIP: wantIP,
	}
	var _ port.MetricsGateway = gateway

	val := 2.0
	err = gateway.MetricsUpdateBatch(context.Background(), []dto.MetricUpdateInput{
		{ID: "Alloc", MType: "gauge", Value: &val},
	})
	require.NoError(t, err)
	assert.Equal(t, wantIP, stub.gotIP)
	assert.Equal(t, 1, stub.gotSize)
}

func TestNewGateway_requiresAddress(t *testing.T) {
	_, err := NewGateway(MetricsGRPCGatewayConfig{})
	require.Error(t, err)
}
