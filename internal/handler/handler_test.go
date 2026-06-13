package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/config"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/repository/mem"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/service"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPing(t *testing.T) {

	storage := mem.NewMemStorage()
	svc := service.NewMetricsService(
		storage,
		nil,
		config.ServerConfig{},
	)
	h := NewMetricsHandler(svc, nil)
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	h.Ping(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

}

func TestGetList(t *testing.T) {

	storage := mem.NewMemStorage()
	require.NoError(t,
		storage.UpdateGauge(t.Context(), "Alloc", 123.45))
	require.NoError(t,
		storage.UpdateCounter(t.Context(), "PollCount", 10))
	svc := service.NewMetricsService(
		storage,
		nil,
		config.ServerConfig{},
	)
	h := NewMetricsHandler(svc, nil)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.GetList(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, "Alloc")
	assert.Contains(t, body, "PollCount")

}

func TestValueGauge(t *testing.T) {

	storage := mem.NewMemStorage()
	require.NoError(t,
		storage.UpdateGauge(t.Context(), "Alloc", 123.45))
	svc := service.NewMetricsService(
		storage,
		nil,
		config.ServerConfig{},
	)
	h := NewMetricsHandler(svc, nil)
	req := httptest.NewRequest(
		http.MethodGet,
		"/value/gauge/Alloc",
		nil,
	)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("type", "gauge")
	rctx.URLParams.Add("name", "Alloc")
	req = req.WithContext(
		context.WithValue(
			req.Context(),
			chi.RouteCtxKey,
			rctx,
		),
	)
	rec := httptest.NewRecorder()
	h.Value(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "123.45", strings.TrimSpace(rec.Body.String()))

}

func TestValueJSONGauge(t *testing.T) {

	storage := mem.NewMemStorage()
	require.NoError(t,
		storage.UpdateGauge(t.Context(), "Alloc", 321.5))
	svc := service.NewMetricsService(
		storage,
		nil,
		config.ServerConfig{},
	)
	h := NewMetricsHandler(svc, nil)
	body := `{
		"id":"Alloc",
		"type":"gauge"
	}`
	req := httptest.NewRequest(
		http.MethodPost,
		"/value",
		bytes.NewBufferString(body),
	)
	rec := httptest.NewRecorder()
	h.ValueJSON(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"id":"Alloc"`)
	assert.Contains(t, rec.Body.String(), `"value":321.5`)

}
func TestValue_InvalidType(t *testing.T) {
	storage := mem.NewMemStorage()

	svc := service.NewMetricsService(
		storage,
		nil,
		config.ServerConfig{},
	)

	h := NewMetricsHandler(svc, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("type", "unknown")
	rctx.URLParams.Add("name", "Alloc")

	req = req.WithContext(
		context.WithValue(req.Context(), chi.RouteCtxKey, rctx),
	)

	rec := httptest.NewRecorder()

	h.Value(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
func TestValue_GaugeNotFound(t *testing.T) {
	storage := mem.NewMemStorage()

	svc := service.NewMetricsService(
		storage,
		nil,
		config.ServerConfig{},
	)

	h := NewMetricsHandler(svc, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("type", "gauge")
	rctx.URLParams.Add("name", "UnknownMetric")

	req = req.WithContext(
		context.WithValue(req.Context(), chi.RouteCtxKey, rctx),
	)

	rec := httptest.NewRecorder()

	h.Value(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestValueJSON_BadJSON(t *testing.T) {
	storage := mem.NewMemStorage()

	svc := service.NewMetricsService(
		storage,
		nil,
		config.ServerConfig{},
	)

	h := NewMetricsHandler(svc, nil)

	req := httptest.NewRequest(
		http.MethodPost,
		"/value",
		strings.NewReader("{invalid"),
	)

	rec := httptest.NewRecorder()

	h.ValueJSON(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestValueJSON_InvalidType(t *testing.T) {
	storage := mem.NewMemStorage()

	svc := service.NewMetricsService(
		storage,
		nil,
		config.ServerConfig{},
	)

	h := NewMetricsHandler(svc, nil)

	body := `{
		"id":"Alloc",
		"type":"unknown"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/value",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	h.ValueJSON(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestValueJSON_EmptyID(t *testing.T) {
	storage := mem.NewMemStorage()

	svc := service.NewMetricsService(
		storage,
		nil,
		config.ServerConfig{},
	)

	h := NewMetricsHandler(svc, nil)

	body := `{
		"id":"",
		"type":"gauge"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/value",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	h.ValueJSON(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
func TestValueJSON_GaugeNotFound(t *testing.T) {
	storage := mem.NewMemStorage()

	svc := service.NewMetricsService(
		storage,
		nil,
		config.ServerConfig{},
	)

	h := NewMetricsHandler(svc, nil)

	body := `{
		"id":"UnknownMetric",
		"type":"gauge"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/value",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	h.ValueJSON(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
func TestValueJSON_CounterSuccess(t *testing.T) {
	storage := mem.NewMemStorage()

	require.NoError(t,
		storage.UpdateCounter(t.Context(), "PollCount", 5))

	svc := service.NewMetricsService(
		storage,
		nil,
		config.ServerConfig{},
	)

	h := NewMetricsHandler(svc, nil)

	body := `{
		"id":"PollCount",
		"type":"counter"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/value",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	h.ValueJSON(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"delta":5`)
}
func TestGetList_MethodNotAllowed(t *testing.T) {
	storage := mem.NewMemStorage()

	svc := service.NewMetricsService(
		storage,
		nil,
		config.ServerConfig{},
	)

	h := NewMetricsHandler(svc, nil)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	h.GetList(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.Equal(t, http.MethodGet, rec.Header().Get("Allow"))
}
