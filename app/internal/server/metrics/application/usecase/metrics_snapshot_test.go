package usecase

import (
	"context"
	"testing"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/port/mocks"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/domain/entity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestMetricsSnapshot_Execute(t *testing.T) {
	ctx := context.Background()
	v := 1.0
	metrics := map[string]entity.Metric{
		"a": {ID: "a", MType: entity.Gauge, Value: &v},
	}

	ctrl := gomock.NewController(t)
	reader := mocks.NewMockMetricReader(ctrl)
	reader.EXPECT().GetAll(ctx).Return(metrics, nil)

	fileRepo := mocks.NewMockMetricFileRepository(ctrl)
	fileRepo.EXPECT().SaveAll(ctx, metrics).Return(nil)

	uc := NewMetricsSnapshot(reader, fileRepo)
	count, err := uc.Execute(ctx, struct{}{})

	require.NoError(t, err)
	assert.Equal(t, 1, count)
}
