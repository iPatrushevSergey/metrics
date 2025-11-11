package handler

import (
	"errors"

	"github.com/iPatrushevSergey/metrics/internal/model"
)

//go:generate easyjson -all metric.go
type MetricDTO struct {
	ID    string      `json:"id"`
	MType string      `json:"type"`
	Value interface{} `json:"value,omitempty"`
	Hash  string      `json:"hash,omitempty"`
}

func modelToDTO(m model.Metric) MetricDTO {
	dto := MetricDTO{
		ID:    m.ID,
		MType: m.MType,
		Hash:  m.Hash,
	}

	switch m.MType {
	case model.Counter:
		if m.Delta != nil {
			dto.Value = *m.Delta
		}
	case model.Gauge:
		if m.Value != nil {
			dto.Value = *m.Value
		}
	}
	return dto
}

func dtoToModel(dto MetricDTO) (model.Metric, error) {
	m := model.Metric{
		ID:    dto.ID,
		MType: dto.MType,
		Hash:  dto.Hash,
	}

	if dto.Value == nil {
		return model.Metric{}, errors.New("the value cannot be empty")
	}

	switch dto.MType {
	case model.Counter:
		f, ok := dto.Value.(float64)
		if !ok {
			return model.Metric{}, errors.New("invalid type for counter value: expected number")
		}
		v := int64(f)
		m.Delta = &v

	case model.Gauge:
		f, ok := dto.Value.(float64)
		if !ok {
			return model.Metric{}, errors.New("invalid type for gauge value: expected number")
		}
		m.Value = &f

	default:
		return model.Metric{}, errors.New("unknown metric type")
	}
	return m, nil
}
