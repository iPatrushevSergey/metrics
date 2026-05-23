package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application"
	"github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/dto"
	portmocks "github.com/iPatrushevSergey/metrics/app/internal/server/metrics/application/port/mocks"
	httpdto "github.com/iPatrushevSergey/metrics/app/internal/server/metrics/presentation/http/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestMetricHandler_PingDB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	log := portmocks.NewMockLogger(ctrl)
	log.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	factory := stubUseCaseFactory{
		pingDB: stubUseCase[struct{}, struct{}]{},
	}
	h := NewMetricHandler(factory, log)

	r := gin.New()
	r.GET("/ping", h.PingDB)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMetricHandler_GetValue(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	log := portmocks.NewMockLogger(ctrl)
	log.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	factory := stubUseCaseFactory{
		getMetricValue: stubUseCase[dto.GetMetricValueInput, string]{
			out: "42",
		},
	}
	h := NewMetricHandler(factory, log)

	r := gin.New()
	r.GET("/value/:type/:name/", h.GetValue)

	req := httptest.NewRequest(http.MethodGet, "/value/gauge/cpu/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "42", w.Body.String())
}

func TestMetricHandler_UpdateJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	log := portmocks.NewMockLogger(ctrl)
	log.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	factory := stubUseCaseFactory{
		upsertMetric: stubUseCase[dto.UpsertMetricInput, struct{}]{},
	}
	h := NewMetricHandler(factory, log)

	r := gin.New()
	r.POST("/update/", h.UpdateJSON)

	body, err := json.Marshal(httpdto.MetricRequest{ID: "cpu", MType: "gauge", Value: ptr(1.0)})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMetricHandler_GetAll(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	log := portmocks.NewMockLogger(ctrl)
	log.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	factory := stubUseCaseFactory{
		getAllMetrics: stubUseCase[struct{}, []dto.MetricForDisplayOutput]{
			out: []dto.MetricForDisplayOutput{{MetricID: "cpu", MetricValue: "42"}},
		},
	}
	h := NewMetricHandler(factory, log)

	r := gin.New()
	r.GET("/", h.GetAll)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "cpu")
}

func TestMetricHandler_Update(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	log := portmocks.NewMockLogger(ctrl)
	log.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	factory := stubUseCaseFactory{
		updateMetric: stubUseCase[dto.UpdateMetricInput, struct{}]{},
	}
	h := NewMetricHandler(factory, log)

	r := gin.New()
	r.POST("/update/:type/:name/:value/", h.Update)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/cpu/42/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMetricHandler_UpdatesJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	log := portmocks.NewMockLogger(ctrl)
	log.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	factory := stubUseCaseFactory{
		upsertMetricsBatch: stubUseCase[dto.UpsertMetricsBatchInput, struct{}]{},
	}
	h := NewMetricHandler(factory, log)

	r := gin.New()
	r.POST("/updates/", h.UpdatesJSON)

	body, err := json.Marshal([]httpdto.MetricRequest{{ID: "cpu", MType: "gauge", Value: ptr(1.0)}})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMetricHandler_GetJSON_success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	log := portmocks.NewMockLogger(ctrl)
	log.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	v := 42.0
	factory := stubUseCaseFactory{
		getMetric: stubUseCase[dto.GetMetricInput, dto.MetricOutput]{
			out: dto.MetricOutput{ID: "cpu", MType: "gauge", Value: &v},
		},
	}
	h := NewMetricHandler(factory, log)

	r := gin.New()
	r.POST("/value/", h.GetJSON)

	body := []byte(`{"id":"cpu","type":"gauge"}`)
	req := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp httpdto.MetricResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "cpu", resp.ID)
}

func TestMetricHandler_GetJSON_notFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	log := portmocks.NewMockLogger(ctrl)
	log.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	factory := stubUseCaseFactory{
		getMetric: stubUseCase[dto.GetMetricInput, dto.MetricOutput]{
			err: application.ErrNotFound,
		},
	}
	h := NewMetricHandler(factory, log)

	r := gin.New()
	r.POST("/value/", h.GetJSON)

	body := []byte(`{"id":"x","type":"gauge"}`)
	req := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func ptr(v float64) *float64 { return &v }
