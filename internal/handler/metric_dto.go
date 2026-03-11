package handler

import (
	"errors"
	"strings"

	"github.com/iPatrushevSergey/metrics/internal/model"
)

//go:generate easyjson -all $GOFILE

// MetricDTO is the metric representation for HTTP JSON API.
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
		MType: string(m.MType),
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
	// Нормализация типа метрики
	normalizedType := strings.ToLower(strings.TrimSpace(dto.MType))

	m := model.Metric{
		ID:    strings.TrimSpace(dto.ID),
		MType: model.MetricType(normalizedType),
		Hash:  strings.TrimSpace(dto.Hash),
	}

	switch model.MetricType(normalizedType) {
	case model.Counter:
		m.Delta = dto.Delta
	case model.Gauge:
		m.Value = dto.Value
	default:
		return model.Metric{}, errors.New("unknown type of metric")
	}
	return m, nil
}
