package usecase

import (
	"context"
	"testing"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/port/mocks"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/domain/entity"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/domain/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetMetric_Execute(t *testing.T) {
	ctx := context.Background()
	v := 42.0
	m := entity.Metric{ID: "cpu", MType: entity.Gauge, Value: &v}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		reader := mocks.NewMockMetricReader(ctrl)
		reader.EXPECT().GetByID(ctx, "cpu").Return(m, nil)

		uc := NewGetMetric(reader, service.MetricService{})
		out, err := uc.Execute(ctx, dto.GetMetricInput{ID: "cpu", MType: "gauge"})

		require.NoError(t, err)
		assert.Equal(t, "cpu", out.ID)
		assert.Equal(t, v, *out.Value)
	})

	t.Run("not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		reader := mocks.NewMockMetricReader(ctrl)
		reader.EXPECT().GetByID(ctx, "cpu").Return(entity.Metric{}, application.ErrNotFound)

		uc := NewGetMetric(reader, service.MetricService{})
		_, err := uc.Execute(ctx, dto.GetMetricInput{ID: "cpu", MType: "gauge"})

		assert.ErrorIs(t, err, application.ErrNotFound)
	})

	t.Run("bad type", func(t *testing.T) {
		uc := NewGetMetric(nil, service.MetricService{})
		_, err := uc.Execute(ctx, dto.GetMetricInput{ID: "x", MType: "unknown"})
		assert.ErrorIs(t, err, application.ErrBadMetricType)
	})

	t.Run("type mismatch", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		v := 1.0
		reader := mocks.NewMockMetricReader(ctrl)
		reader.EXPECT().GetByID(ctx, "a").Return(entity.Metric{ID: "a", MType: entity.Gauge, Value: &v}, nil)

		uc := NewGetMetric(reader, service.MetricService{})
		_, err := uc.Execute(ctx, dto.GetMetricInput{ID: "a", MType: "counter"})
		assert.ErrorIs(t, err, application.ErrNotFound)
	})
}
