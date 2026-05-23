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

func TestGetMetricValue_Execute(t *testing.T) {
	ctx := context.Background()
	v := 3.14
	m := entity.Metric{ID: "cpu", MType: entity.Gauge, Value: &v}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		reader := mocks.NewMockMetricReader(ctrl)
		reader.EXPECT().GetByID(ctx, "cpu").Return(m, nil)

		uc := NewGetMetricValue(reader, service.MetricService{})
		got, err := uc.Execute(ctx, dto.GetMetricValueInput{ID: "cpu", MType: "gauge"})

		require.NoError(t, err)
		assert.Equal(t, "3.14", got)
	})

	t.Run("not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		reader := mocks.NewMockMetricReader(ctrl)
		reader.EXPECT().GetByID(ctx, "x").Return(entity.Metric{}, application.ErrNotFound)

		uc := NewGetMetricValue(reader, service.MetricService{})
		_, err := uc.Execute(ctx, dto.GetMetricValueInput{ID: "x", MType: "gauge"})
		assert.ErrorIs(t, err, application.ErrNotFound)
	})
}
