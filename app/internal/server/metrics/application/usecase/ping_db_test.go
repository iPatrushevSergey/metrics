package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/port/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestPingDB_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("ok", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		reader := mocks.NewMockMetricReader(ctrl)
		reader.EXPECT().Ping(ctx).Return(nil)

		uc := NewPingDB(reader)
		_, err := uc.Execute(ctx, struct{}{})
		assert.NoError(t, err)
	})

	t.Run("fail", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		reader := mocks.NewMockMetricReader(ctrl)
		reader.EXPECT().Ping(ctx).Return(errors.New("db down"))

		uc := NewPingDB(reader)
		_, err := uc.Execute(ctx, struct{}{})
		assert.ErrorIs(t, err, application.ErrInternal)
	})
}
