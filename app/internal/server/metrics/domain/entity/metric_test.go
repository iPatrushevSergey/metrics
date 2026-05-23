package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetric_gauge(t *testing.T) {
	m, err := NewMetric(Gauge, "cpu", "1.5")
	require.NoError(t, err)
	assert.Equal(t, 1.5, *m.Value)
}

func TestNewMetric_invalidValue(t *testing.T) {
	_, err := NewMetric(Gauge, "cpu", "bad")
	assert.Error(t, err)
}

func TestMetric_ValidateMetricValues(t *testing.T) {
	v := 1.0
	assert.NoError(t, (Metric{MType: Gauge, Value: &v}).ValidateMetricValues())

	d := int64(1)
	assert.NoError(t, (Metric{MType: Counter, Delta: &d}).ValidateMetricValues())

	assert.Error(t, (Metric{MType: Gauge}).ValidateMetricValues())
}
