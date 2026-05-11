// Package service holds metric helpers outside entity state mutations (compare balance MetricService ↔ entity methods).
package service

import (
	"fmt"
	"strings"

	"github.com/iPatrushevSergey/metrics/metrics_new/app/internal/server/metrics/domain/entity"
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
