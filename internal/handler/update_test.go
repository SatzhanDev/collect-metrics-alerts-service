package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

type mockMetricsService struct {
	gauges   map[string]float64
	counters map[string]int64
}

func newMockMetricsService() *mockMetricsService {
	return &mockMetricsService{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (m *mockMetricsService) UpdateGauge(name string, v float64) error {
	m.gauges[name] = v
	return nil
}

func (m *mockMetricsService) UpdateCounter(name string, delta int64) error {
	m.counters[name] += delta
	return nil
}

func TestMetricsHandler_Update_MethodNotAllowed(t *testing.T) {
	svc := newMockMetricsService()
	h := NewMetricsHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/update/gauge/Alloc/1", nil)
	rr := httptest.NewRecorder()

	h.Update(rr, req)

	require.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestMetricsHandler_Update_NotFound_WhenMissingParts(t *testing.T) {
	svc := newMockMetricsService()
	h := NewMetricsHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge", nil) // мало сегментов
	rr := httptest.NewRecorder()

	h.Update(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)

	// if rr.Code != http.StatusNotFound {
	// 	t.Fatalf("expected %d, got %d", http.StatusNotFound, rr.Code)
	// }
}

func TestMetricsHandler_Update_BadRequest_WhenInvalidType(t *testing.T) {
	svc := newMockMetricsService()
	h := NewMetricsHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/update/unknown/Alloc/1", nil)
	rr := httptest.NewRecorder()

	h.Update(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)

	// if rr.Code != http.StatusBadRequest {
	// 	t.Fatalf("expected %d, got %d", http.StatusBadRequest, rr.Code)
	// }

}
func TestMetricsHandler_Update_BadRequest_WhenGaugeValueInvalid(t *testing.T) {
	svc := newMockMetricsService()
	h := NewMetricsHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/Alloc/not-a-number", nil)
	rr := httptest.NewRecorder()

	h.Update(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)

	// if rr.Code != http.StatusBadRequest {
	// 	t.Fatalf("expected %d, got %d", http.StatusBadRequest, rr.Code)
	// }
}

func TestMetricsHandler_Update_OK_Gauge(t *testing.T) {
	svc := newMockMetricsService()
	h := NewMetricsHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/Alloc/12.5", nil)
	rr := httptest.NewRecorder()

	h.Update(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	// if rr.Code != http.StatusOK {
	// 	t.Fatalf("expected %d, got %d", http.StatusOK, rr.Code)
	// }

	require.Equal(t, 12.5, svc.gauges["Alloc"])
	// if got := svc.gauges["Alloc"]; got != 12.5 {
	// 	t.Fatalf("expected service to store Alloc=12.5, got %v", got)
	// }
}

func TestMetricsHandler_Update_OK_Counter(t *testing.T) {
	svc := newMockMetricsService()
	h := NewMetricsHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/update/counter/PollCount/3", nil)
	rr := httptest.NewRecorder()

	h.Update(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	// if rr.Code != http.StatusOK {
	// 	t.Fatalf("expected %d, got %d", http.StatusOK, rr.Code)
	// }

	require.Equal(t, int64(3), svc.counters["PollCount"])
	// if got := svc.counters["PollCount"]; got != 3 {
	// 	t.Fatalf("expected service to store PollCount+=3, got %v", got)
	// }
}
