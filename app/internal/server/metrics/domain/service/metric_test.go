package service

import (
	"testing"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricService_CheckMetricType(t *testing.T) {
	svc := MetricService{}

	t.Run("gauge", func(t *testing.T) {
		mType, err := svc.CheckMetricType("gauge")
		require.NoError(t, err)
		assert.Equal(t, entity.Gauge, mType)
	})

	t.Run("invalid", func(t *testing.T) {
		_, err := svc.CheckMetricType("unknown")
		assert.Error(t, err)
	})
}

func TestMetricService_MergeMetricsByID(t *testing.T) {
	svc := MetricService{}
	v1, v2 := 1.0, 2.0

	merged, err := svc.MergeMetricsByID([]entity.Metric{
		{ID: "a", MType: entity.Gauge, Value: &v1},
		{ID: "a", MType: entity.Gauge, Value: &v2},
	})
	require.NoError(t, err)
	require.Len(t, merged, 1)
	assert.Equal(t, v2, *merged[0].Value)
}

func TestMetricService_FormatMetricsValue(t *testing.T) {
	svc := MetricService{}
	v := 42.0

	out, err := svc.FormatMetricsValue(map[string]entity.Metric{
		"cpu": {ID: "cpu", MType: entity.Gauge, Value: &v},
	})
	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Equal(t, "cpu", out[0].ID)
}

func TestMetricService_CollectIDs(t *testing.T) {
	svc := MetricService{}
	ids := svc.CollectIDs([]entity.Metric{{ID: "a"}, {ID: "b"}})
	assert.Equal(t, []string{"a", "b"}, ids)
}
