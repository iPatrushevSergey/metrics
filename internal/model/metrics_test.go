package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMetricType_values(t *testing.T) {
	require.Equal(t, MetricType("counter"), Counter)
	require.Equal(t, MetricType("gauge"), Gauge)
}
