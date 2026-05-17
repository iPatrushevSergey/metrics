package model

// Metric is the DB projection of the metrics table row.
type Metric struct {
	ID    string
	MType string
	Delta *int64
	Value *float64
	Hash  *string
}
