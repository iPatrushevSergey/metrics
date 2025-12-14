package repository

import (
	"context"
	"errors"

	"github.com/iPatrushevSergey/metrics/internal/model"
)

var (
	// ErrNotFound возвращается, когда метрика не найдена в хранилище
	ErrNotFound = errors.New("metric not found")
	// ErrAlreadyExists возвращается, когда попытка создать метрику с уже существующим ID
	ErrAlreadyExists = errors.New("metric already exists")
)

// MetricRepository определяет интерфейс для работы с хранилищем метрик
type MetricRepository interface {
	// GetByID возвращает метрику по идентификатору
	GetByID(ctx context.Context, id string) (model.Metric, error)
	// GetAll возвращает все метрики из хранилища
	GetAll(ctx context.Context) (map[string]model.Metric, error)
	// Create создает новую метрику в хранилище
	Create(ctx context.Context, metric model.Metric) error
	// Update обновляет существующую метрику по идентификатору
	Update(ctx context.Context, id string, metric model.Metric) error
	// Ping проверяет доступность хранилища
	Ping(ctx context.Context) error
}
