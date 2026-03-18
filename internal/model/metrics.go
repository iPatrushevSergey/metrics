// Package model defines domain types for metrics (gauge, counter).
package model

// MetricType is the type of metric (gauge or counter).
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
	Hash  string
}
