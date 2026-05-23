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

func TestNewMetric_counter(t *testing.T) {
	m, err := NewMetric(Counter, "hits", "3")
	require.NoError(t, err)
	assert.Equal(t, int64(3), *m.Delta)
}

func TestMetric_FormatValueAsString(t *testing.T) {
	v := 1.5
	s, err := (Metric{MType: Gauge, Value: &v}).FormatValueAsString()
	require.NoError(t, err)
	assert.Equal(t, "1.5", s)

	d := int64(7)
	s, err = (Metric{MType: Counter, Delta: &d}).FormatValueAsString()
	require.NoError(t, err)
	assert.Equal(t, "7", s)
}

func TestMetric_ApplyUpdate_counter(t *testing.T) {
	d1, d2 := int64(1), int64(2)
	m := Metric{MType: Counter, Delta: &d1}
	require.NoError(t, m.ApplyUpdate(Metric{MType: Counter, Delta: &d2}))
	assert.Equal(t, int64(3), *m.Delta)
}

func TestMetric_MatchMetricTypes(t *testing.T) {
	m := Metric{MType: Gauge}
	assert.NoError(t, m.MatchMetricTypes(Gauge))
	assert.Error(t, m.MatchMetricTypes(Counter))
}

func TestMetric_FormatValueAsString_errors(t *testing.T) {
	_, err := (Metric{MType: Gauge}).FormatValueAsString()
	assert.Error(t, err)

	_, err = (Metric{MType: Counter}).FormatValueAsString()
	assert.Error(t, err)

	_, err = (Metric{MType: "unknown"}).FormatValueAsString()
	assert.Error(t, err)
}

func TestMetric_ApplyUpdate_unsupported(t *testing.T) {
	m := Metric{MType: Gauge}
	assert.Error(t, m.ApplyUpdate(Metric{MType: "bad"}))
}
