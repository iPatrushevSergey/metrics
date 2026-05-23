package converter

import (
	"testing"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricConverter_roundTrip(t *testing.T) {
	c := MetricConverterImpl{}
	v := 1.5
	m := entity.Metric{ID: "a", MType: entity.Gauge, Value: &v, Hash: "h"}

	row := c.ToModel(m)
	got := c.ToEntity(row)
	assert.Equal(t, m.ID, got.ID)
	assert.Equal(t, m.MType, got.MType)
	require.NotNil(t, got.Value)
	assert.Equal(t, v, *got.Value)
	assert.Equal(t, "h", got.Hash)
	assert.Equal(t, string(entity.Gauge), row.MType)
}
