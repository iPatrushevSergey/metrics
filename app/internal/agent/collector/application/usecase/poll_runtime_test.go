package usecase

import (
	"context"
	"runtime"
	"testing"

	portmocks "github.com/iPatrushevSergey/metrics/app/internal/agent/collector/application/port/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestPollRuntimeTick_Execute(t *testing.T) {
	ctx := context.Background()
	ms := runtime.MemStats{Alloc: 42}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		sampler := portmocks.NewMockMetricsSampler(ctrl)
		repo := portmocks.NewMockMetricsRepository(ctrl)

		sampler.EXPECT().ReadMemStats().Return(ms)
		repo.EXPECT().UpdateRuntimeMetrics(ms, 3.14)

		uc := NewPollRuntimeTick(sampler, repo, func() float64 { return 3.14 })
		n, err := uc.Execute(ctx, struct{}{})

		assert.NoError(t, err)
		assert.Equal(t, 1, n)
	})

	t.Run("context canceled", func(t *testing.T) {
		cctx, cancel := context.WithCancel(context.Background())
		cancel()

		uc := NewPollRuntimeTick(nil, nil, nil)
		n, err := uc.Execute(cctx, struct{}{})

		assert.ErrorIs(t, err, context.Canceled)
		assert.Equal(t, 0, n)
	})
}
