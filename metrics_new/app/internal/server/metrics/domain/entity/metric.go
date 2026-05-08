// Package entity defines domain types for metrics (gauge, counter).
package entity

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
