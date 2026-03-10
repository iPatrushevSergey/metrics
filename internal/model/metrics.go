package model

const (
	Counter = "counter"
	Gauge   = "gauge"
)

// Metric is a single metric.
type Metric struct {
	ID    string
	MType string
	Delta *int64
	Value *float64
	Hash  string
}
