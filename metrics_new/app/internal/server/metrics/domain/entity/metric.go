// Package holds metric entities.
package entity

import (
	"fmt"
	"strconv"
	"strings"
)

// MetricType is the type of metric.
type MetricType string

const (
	Counter MetricType = "counter"
	Gauge   MetricType = "gauge"
)

// Metric is a single metric.
type Metric struct {
	ID    string
	MType MetricType
	Delta *int64
	Value *float64
}

// NewMetric builds a metric.
func NewMetric(mType MetricType, id, value string) (Metric, error) {
	m := Metric{ID: strings.TrimSpace(id), MType: mType}
	switch mType {
	case Gauge:
		v, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil {
			return Metric{}, fmt.Errorf("parse gauge: %w", err)
		}
		m.Value = &v
	case Counter:
		v, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err != nil {
			return Metric{}, fmt.Errorf("parse counter: %w", err)
		}
		m.Delta = &v
	default:
		return Metric{}, fmt.Errorf("unknown metric MType: %s", mType)
	}
	return m, nil
}

// ValidateMetricValues checks that Delta/Value are consistent with MType.
func (m Metric) ValidateMetricValues() error {
	switch m.MType {
	case Counter:
		if m.Delta == nil {
			return ErrMissingCounterDelta
		}
	case Gauge:
		if m.Value == nil {
			return ErrMissingGaugeValue
		}
	default:
		return ErrUnsupportedMetricType
	}
	return nil
}

// ApplyUpdate merges an incoming metric into the receiver.
// Counter: sum non-nil deltas; gauge: replace value.
func (m *Metric) ApplyUpdate(incoming Metric) error {
	if m.ID != incoming.ID || m.MType != incoming.MType {
		return ErrMetricIdentityMismatch
	}
	switch incoming.MType {
	case Counter:
		if incoming.Delta != nil {
			if m.Delta != nil {
				sum := *m.Delta + *incoming.Delta
				m.Delta = &sum
			} else {
				m.Delta = incoming.Delta
			}
		}
	case Gauge:
		m.Value = incoming.Value
	default:
		return ErrUnsupportedMetricType
	}
	return nil
}

// FormatValueAsString returns the representation of the metric value as a string.
func (m Metric) FormatValueAsString() (string, error) {
	switch m.MType {
	case Gauge:
		if m.Value == nil {
			return "", fmt.Errorf("gauge value is missing")
		}
		return strconv.FormatFloat(*m.Value, 'f', -1, 64), nil
	case Counter:
		if m.Delta == nil {
			return "", fmt.Errorf("counter delta is missing")
		}
		return strconv.FormatInt(*m.Delta, 10), nil
	default:
		return "", fmt.Errorf("unsupported metric type: %s", m.MType)
	}
}

// MergeMetricsByID merges metrics by ID preserving first-seen order.
func MergeMetricsByID(metrics []Metric) ([]Metric, error) {
	idToMetric := make(map[string]Metric, len(metrics))
	metricIDs := make([]string, 0, len(metrics))

	for _, metric := range metrics {
		id := metric.ID
		prevMetric, exists := idToMetric[id]
		if !exists {
			idToMetric[id] = metric
			metricIDs = append(metricIDs, id)
			continue
		}
		if err := prevMetric.ApplyUpdate(metric); err != nil {
			return nil, fmt.Errorf("merge metric %q: %w", id, err)
		}
		idToMetric[id] = prevMetric
	}

	updatedMetrics := make([]Metric, 0, len(idToMetric))
	for _, id := range metricIDs {
		updatedMetrics = append(updatedMetrics, idToMetric[id])
	}
	return updatedMetrics, nil
}
