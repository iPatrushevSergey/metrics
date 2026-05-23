package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/dto"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/port"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/presentation/factory"
	"github.com/iPatrushevSergey/metrics/app/internal/pkg/adapters/logger"
	"github.com/stretchr/testify/assert"
)

type stubUseCase[In, Out any] struct{}

func (stubUseCase[In, Out]) Execute(context.Context, In) (Out, error) {
	var zero Out
	return zero, nil
}

type stubFactory struct{}

func (stubFactory) GetMetricValueUseCase() port.UseCase[dto.GetMetricValueInput, string] {
	return stubUseCase[dto.GetMetricValueInput, string]{}
}
func (stubFactory) GetMetricUseCase() port.UseCase[dto.GetMetricInput, dto.MetricOutput] {
	return stubUseCase[dto.GetMetricInput, dto.MetricOutput]{}
}
func (stubFactory) UpdateMetricUseCase() port.UseCase[dto.UpdateMetricInput, struct{}] {
	return stubUseCase[dto.UpdateMetricInput, struct{}]{}
}
func (stubFactory) UpsertMetricUseCase() port.UseCase[dto.UpsertMetricInput, struct{}] {
	return stubUseCase[dto.UpsertMetricInput, struct{}]{}
}
func (stubFactory) UpsertMetricsBatchUseCase() port.UseCase[dto.UpsertMetricsBatchInput, struct{}] {
	return stubUseCase[dto.UpsertMetricsBatchInput, struct{}]{}
}
func (stubFactory) GetAllMetricsUseCase() port.UseCase[struct{}, []dto.MetricForDisplayOutput] {
	return stubUseCase[struct{}, []dto.MetricForDisplayOutput]{}
}
func (stubFactory) PingDBUseCase() port.UseCase[struct{}, struct{}] { return stubUseCase[struct{}, struct{}]{} }
func (stubFactory) MetricsSnapshotUseCase() port.UseCase[struct{}, int] {
	return stubUseCase[struct{}, int]{}
}
func (stubFactory) RestoreMetricsFromFileUseCase() port.UseCase[struct{}, struct{}] {
	return stubUseCase[struct{}, struct{}]{}
}
func (stubFactory) RecordAuditToFileUseCase() port.UseCase[dto.AuditEvent, struct{}] {
	return stubUseCase[dto.AuditEvent, struct{}]{}
}
func (stubFactory) CreateRemoteAuditUseCase() port.UseCase[dto.AuditEvent, struct{}] {
	return stubUseCase[dto.AuditEvent, struct{}]{}
}

var _ factory.UseCaseFactory = stubFactory{}

func TestRegisterRoutes_ping(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterRoutes(r, stubFactory{}, logger.NewNopLogger())

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
