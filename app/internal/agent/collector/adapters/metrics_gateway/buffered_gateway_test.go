package metricsgateway

import (
	"context"
	"testing"
	"time"

	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/application/dto"
	"github.com/iPatrushevSergey/metrics/app/internal/agent/collector/application/port/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewBufferedMetricsGateway_invalidWorkers(t *testing.T) {
	_, err := NewBufferedMetricsGateway(nil, nil, 0, context.Background())
	assert.Error(t, err)
}

func TestBufferedMetricsGateway_MetricsUpdateBatch(t *testing.T) {
	ctrl := gomock.NewController(t)

	done := make(chan struct{})
	delegate := mocks.NewMockMetricsGateway(ctrl)
	delegate.EXPECT().MetricsUpdateBatch(gomock.Any(), gomock.Any()).DoAndReturn(
		func(context.Context, []dto.MetricUpdateInput) error {
			close(done)
			return nil
		},
	)

	log := mocks.NewMockLogger(ctrl)
	sendCtx := context.Background()

	bg, err := NewBufferedMetricsGateway(delegate, log, 1, sendCtx)
	require.NoError(t, err)
	bg.Start()
	t.Cleanup(bg.Stop)

	err = bg.MetricsUpdateBatch(context.Background(), nil)
	require.NoError(t, err)

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("delegate was not called")
	}
}
