package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	models "github.com/SatzhanDev/collect-metrics-alerts-service/internal/model"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/require"
)

type mockMetricsService struct {
	gauges   map[string]float64
	counters map[string]int64
	err      error
}

func newMockMetricsService() *mockMetricsService {
	return &mockMetricsService{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (m *mockMetricsService) UpdateGauge(ctx context.Context, name string, v float64) error {
	m.gauges[name] = v
	return nil
}

func (m *mockMetricsService) UpdateCounter(ctx context.Context, name string, delta int64) error {
	m.counters[name] += delta
	return nil
}
func (m *mockMetricsService) GetGauge(ctx context.Context, name string) (float64, error) {
	v, ok := m.gauges[name]
	if !ok {
		return 0, errors.New("metric not found")
	}
	return v, nil
}
func (m *mockMetricsService) GetCounter(ctx context.Context, name string) (int64, error) {
	v, ok := m.counters[name]
	if !ok {
		return 0, errors.New("metric not found")
	}
	return v, nil
}
func (m *mockMetricsService) GetAll(ctx context.Context) (map[string]float64, map[string]int64, error) {
	if m.err != nil {
		return nil, nil, m.err
	}
	return m.gauges, m.counters, nil
}

func (m *mockMetricsService) RestoreFromFile(ctx context.Context) error {
	return nil
}
func (m *mockMetricsService) SaveToFile(ctx context.Context) error {
	return nil
}
func (m *mockMetricsService) UpdateBatch(ctx context.Context, metrics []models.Metrics) error {
	return nil
}
func (m *mockMetricsService) Ping(ctx context.Context) error {
	return nil
}
func TestMetricsHandler_Update_MethodNotAllowed(t *testing.T) {
	svc := newMockMetricsService()
	h := NewMetricsHandler(svc)

	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", h.Update)

	req := httptest.NewRequest(http.MethodGet, "/update/gauge/Alloc/1", nil)
	rr := httptest.NewRecorder()

	// h.Update(rr, req)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestMetricsHandler_Update_NotFound_WhenMissingParts(t *testing.T) {
	svc := newMockMetricsService()
	h := NewMetricsHandler(svc)

	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", h.Update)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge", nil)
	rr := httptest.NewRecorder()

	// h.Update(rr, req)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)

}

func TestMetricsHandler_Update_BadRequest_WhenInvalidType(t *testing.T) {
	svc := newMockMetricsService()
	h := NewMetricsHandler(svc)

	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", h.Update)

	req := httptest.NewRequest(http.MethodPost, "/update/unknown/Alloc/1", nil)
	rr := httptest.NewRecorder()

	// h.Update(rr, req)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)

}
func TestMetricsHandler_Update_BadRequest_WhenGaugeValueInvalid(t *testing.T) {
	svc := newMockMetricsService()
	h := NewMetricsHandler(svc)

	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", h.Update)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/Alloc/not-a-number", nil)
	rr := httptest.NewRecorder()

	// h.Update(rr, req)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)

}

func TestMetricsHandler_Update_OK_Gauge(t *testing.T) {
	svc := newMockMetricsService()
	h := NewMetricsHandler(svc)

	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", h.Update)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/Alloc/12.5", nil)
	rr := httptest.NewRecorder()

	// h.Update(rr, req)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	require.Equal(t, 12.5, svc.gauges["Alloc"])

}

func TestMetricsHandler_Update_OK_Counter(t *testing.T) {
	svc := newMockMetricsService()
	h := NewMetricsHandler(svc)

	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", h.Update)

	req := httptest.NewRequest(http.MethodPost, "/update/counter/PollCount/3", nil)
	rr := httptest.NewRecorder()

	// h.Update(rr, req)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	require.Equal(t, int64(3), svc.counters["PollCount"])

}
