// Package grpc implements the metrics gRPC API.
package grpc

import (
	"context"
	"errors"
	"strings"

	metricspb "github.com/iPatrushevSergey/metrics/app/internal/pkg/grpc/metrics"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/netutil"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application"
	appdto "github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/port"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/presentation/factory"
	grpcdto "github.com/iPatrushevSergey/metrics/app/internal/server/metrics/presentation/grpc/dto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// MetricsService implements the metrics gRPC service.
type MetricsService struct {
	metricspb.UnimplementedMetricsServer
	useCases factory.UseCaseFactory
	log      port.Logger
}

// NewMetricsService creates a new metrics gRPC service.
func NewMetricsService(useCases factory.UseCaseFactory, log port.Logger) *MetricsService {
	return &MetricsService{useCases: useCases, log: log}
}

// UpdateMetrics handles a batch of metrics updates.
func (s *MetricsService) UpdateMetrics(
	ctx context.Context,
	req *metricspb.UpdateMetricsRequest,
) (*metricspb.UpdateMetricsResponse, error) {
	clientIP := ""
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get(netutil.RealIPHeader); len(vals) > 0 {
			clientIP = vals[0]
		}
	}

	var reqDTOs []grpcdto.Metric
	if req != nil && len(req.GetMetrics()) > 0 {
		reqDTOs = make([]grpcdto.Metric, 0, len(req.GetMetrics()))
		for _, m := range req.GetMetrics() {
			item := grpcdto.Metric{ID: strings.TrimSpace(m.GetId())}
			switch m.GetType() {
			case metricspb.Metric_COUNTER:
				item.MType = "counter"
				d := m.GetDelta()
				item.Delta = &d
			default:
				item.MType = "gauge"
				v := m.GetValue()
				item.Value = &v
			}
			reqDTOs = append(reqDTOs, item)
		}
	}

	inDTOs := make([]appdto.UpsertMetricInput, 0, len(reqDTOs))
	for _, reqDTO := range reqDTOs {
		if strings.TrimSpace(reqDTO.ID) == "" {
			return nil, status.Error(codes.InvalidArgument, "the metric name is missing")
		}
		inDTOs = append(inDTOs, appdto.UpsertMetricInput{
			ID:    reqDTO.ID,
			MType: reqDTO.MType,
			Delta: reqDTO.Delta,
			Value: reqDTO.Value,
		})
	}

	if _, err := s.useCases.UpsertMetricsBatchUseCase().Execute(ctx, appdto.UpsertMetricsBatchInput{
		Metrics:   inDTOs,
		IPAddress: clientIP,
	}); err != nil {
		switch {
		case errors.Is(err, application.ErrBadMetricType), errors.Is(err, application.ErrBadMetricValue):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			s.log.Error("grpc update metrics batch failed", "error", err)
			return nil, status.Error(codes.Internal, application.ErrInternal.Error())
		}
	}

	return &metricspb.UpdateMetricsResponse{}, nil
}
