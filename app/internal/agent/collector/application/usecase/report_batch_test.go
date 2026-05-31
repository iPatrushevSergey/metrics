package usecase

import (
	"context"
	"errors"
	"runtime"
	"testing"
	"time"

	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/adapters/repository/inmemory"
	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/application/port/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestReportBatchTick_Execute(t *testing.T) {
	ctx := context.Background()
	interval := time.Second

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		repo := inmemory.NewMetricsRepository()
		repo.UpdateRuntimeMetrics(runtime.MemStats{Alloc: 1}, 2.5)
		repo.UpdateGopsutilMetrics(100, 50, []float64{5})

		gateway := mocks.NewMockMetricsGateway(ctrl)
		gateway.EXPECT().MetricsUpdateBatch(ctx, gomock.Any()).Return(nil)

		log := mocks.NewMockLogger(ctrl)
		log.EXPECT().Debug("batch sent", "count", gomock.Any())

		uc := NewReportBatchTick(repo, gateway, log, interval)
		n, err := uc.Execute(ctx, struct{}{})

		assert.NoError(t, err)
		assert.Greater(t, n, 0)
	})

	t.Run("gateway error", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		repo := inmemory.NewMetricsRepository()
		repo.UpdateRuntimeMetrics(runtime.MemStats{Alloc: 1}, 1)

		gateway := mocks.NewMockMetricsGateway(ctrl)
		gateway.EXPECT().MetricsUpdateBatch(ctx, gomock.Any()).Return(errors.New("network"))

		uc := NewReportBatchTick(repo, gateway, nil, interval)
		_, err := uc.Execute(ctx, struct{}{})

		assert.Error(t, err)
	})

	t.Run("context canceled", func(t *testing.T) {
		cctx, cancel := context.WithCancel(context.Background())
		cancel()

		uc := NewReportBatchTick(nil, nil, nil, interval)
		n, err := uc.Execute(cctx, struct{}{})

		assert.ErrorIs(t, err, context.Canceled)
		assert.Equal(t, 0, n)
	})
}
