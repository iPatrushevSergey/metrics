package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/application/port/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestPollGopsutilTick_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		sampler := mocks.NewMockMetricsSampler(ctrl)
		repo := mocks.NewMockMetricsRepository(ctrl)

		sampler.EXPECT().ReadVirtualMemory().Return(uint64(1000), uint64(500), nil)
		sampler.EXPECT().ReadCPUPercent(time.Second, true).Return([]float64{10, 20}, nil)
		repo.EXPECT().UpdateGopsutilMetrics(float64(1000), float64(500), []float64{10, 20})

		uc := NewPollGopsutilTick(sampler, repo, nil)
		n, err := uc.Execute(ctx, struct{}{})

		assert.NoError(t, err)
		assert.Equal(t, 1, n)
	})

	t.Run("memory read error", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		sampler := mocks.NewMockMetricsSampler(ctrl)
		log := mocks.NewMockLogger(ctrl)

		sampler.EXPECT().ReadVirtualMemory().Return(uint64(0), uint64(0), errors.New("mem"))
		log.EXPECT().Error("memory metrics", "error", gomock.Any())

		uc := NewPollGopsutilTick(sampler, nil, log)
		n, err := uc.Execute(ctx, struct{}{})

		assert.NoError(t, err)
		assert.Equal(t, 0, n)
	})

	t.Run("context canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		uc := NewPollGopsutilTick(nil, nil, nil)
		n, err := uc.Execute(ctx, struct{}{})

		assert.ErrorIs(t, err, context.Canceled)
		assert.Equal(t, 0, n)
	})
}
