package handler

import (
	"errors"

	"github.com/iPatrushevSergey/metrics/internal/model"
)

//go:generate easyjson -all $GOFILE
type MetricDTO struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

func modelToDTO(m model.Metric) MetricDTO {
	dto := MetricDTO{
		ID:    m.ID,
		MType: m.MType,
		Hash:  m.Hash,
	}

	switch m.MType {
	case model.Counter:
		dto.Delta = m.Delta
	case model.Gauge:
		dto.Value = m.Value
	}
	return dto
}

func dtoToModel(dto MetricDTO) (model.Metric, error) {
	m := model.Metric{
		ID:    dto.ID,
		MType: dto.MType,
		Hash:  dto.Hash,
	}

	switch dto.MType {
	case model.Counter:
		m.Delta = dto.Delta
	case model.Gauge:
		m.Value = dto.Value
	default:
		return model.Metric{}, errors.New("unknown type of metric")
	}
	return m, nil
}
