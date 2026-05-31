// Package dto holds gRPC request and response shapes for the metrics API.
package dto

// Metric is a single metric update from a client request.
type Metric struct {
	ID    string
	MType string
	Delta *int64
	Value *float64
}
