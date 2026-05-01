package dto

// MetricUpdateInput is an application-level input payload.
type MetricUpdateInput struct {
	ID    string
	MType string
	Delta *int64
	Value *float64
}
