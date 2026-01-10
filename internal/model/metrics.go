package model

const (
	Counter = "counter"
	Gauge   = "gauge"
)

// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metric struct {
	ID    string
	MType string
	Delta *int64
	Value *float64
	Hash  string
}
