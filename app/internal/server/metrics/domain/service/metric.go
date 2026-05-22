// Package service holds metric domain logic.
package service

import (
	"fmt"
	"sort"
	"strings"

	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/domain/entity"
)

// MetricService performs operations on metrics.
type MetricService struct{}

// CheckMetricType normalizes and validates an external metric type label.
func (MetricService) CheckMetricType(s string) (entity.MetricType, error) {
	mType := entity.MetricType(strings.ToLower(strings.TrimSpace(s)))
	switch mType {
	case entity.Counter, entity.Gauge:
		return mType, nil
	default:
		return "", fmt.Errorf("unknown metric type")
	}
}

// MergeMetricsByID merges metrics by ID preserving first-seen order.
func (MetricService) MergeMetricsByID(metrics []entity.Metric) ([]entity.Metric, error) {
	idToMetric := make(map[string]entity.Metric, len(metrics))
	metricIDs := make([]string, 0, len(metrics))

	for _, metric := range metrics {
		id := metric.ID
		prevMetric, exists := idToMetric[id]
		if !exists {
			idToMetric[id] = metric
			metricIDs = append(metricIDs, id)
			continue
		}
		if err := prevMetric.MatchMetricTypes(metric.MType); err != nil {
			return nil, fmt.Errorf("merge metric %q: %w", id, err)
		}
		if err := prevMetric.ApplyUpdate(metric); err != nil {
			return nil, fmt.Errorf("merge metric %q: %w", id, err)
		}
		idToMetric[id] = prevMetric
	}

	updatedMetrics := make([]entity.Metric, 0, len(idToMetric))
	for _, id := range metricIDs {
		updatedMetrics = append(updatedMetrics, idToMetric[id])
	}
	return updatedMetrics, nil
}

// CollectIDs returns the IDs of the provided metrics in order.
func (MetricService) CollectIDs(metrics []entity.Metric) []string {
	ids := make([]string, len(metrics))
	for i, m := range metrics {
		ids[i] = m.ID
	}
	return ids
}

// FormatMetricsValue collects metrics from the map, orders rows by ID ascending, and formats each value.
func (MetricService) FormatMetricsValue(idToMetric map[string]entity.Metric) ([]entity.MetricWithValue, error) {
	ids := make([]string, 0, len(idToMetric))
	for id := range idToMetric {
		ids = append(ids, id)
	}

	sort.Strings(ids)

	metricsWithValue := make([]entity.MetricWithValue, 0, len(ids))
	for _, id := range ids {
		metric := idToMetric[id]
		formattedValue, err := metric.FormatValueAsString()
		if err != nil {
			return nil, fmt.Errorf("format metric %q: %w", id, err)
		}
		metricsWithValue = append(metricsWithValue, entity.MetricWithValue{ID: id, FormattedValue: formattedValue})
	}
	return metricsWithValue, nil
}

// BuildCreateUpdateBatches builds the lists of metrics to create and update.
func (MetricService) BuildCreateUpdateBatches(
	existingIDToMetric map[string]entity.Metric,
	mergedMetrics []entity.Metric,
) ([]entity.Metric, []entity.Metric, error) {
	var creates []entity.Metric
	var updates []entity.Metric

	for _, metric := range mergedMetrics {
		existingMetric, exists := existingIDToMetric[metric.ID]
		if !exists {
			creates = append(creates, metric)
			continue
		}
		if err := existingMetric.MatchMetricTypes(metric.MType); err != nil {
			return nil, nil, err
		}
		if err := existingMetric.ApplyUpdate(metric); err != nil {
			return nil, nil, err
		}
		updates = append(updates, existingMetric)
	}
	return creates, updates, nil
}
