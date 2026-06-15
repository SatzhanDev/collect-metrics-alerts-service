package handler

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/config"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/repository/mem"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/service"
	"github.com/go-chi/chi"
)

// ExampleMetricsHandler_Update демонстрирует обновление gauge-метрики через URL-параметры.
func ExampleMetricsHandler_Update() {
	storage := mem.NewMemStorage()
	svc := service.NewMetricsService(storage, nil, config.ServerConfig{})
	h := NewMetricsHandler(svc, nil)

	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", h.Update)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/Alloc/12.5", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	fmt.Println(rec.Code)

	// Output:
	// 200
}

// ExampleMetricsHandler_UpdateJSON демонстрирует обновление counter-метрики через JSON.
func ExampleMetricsHandler_UpdateJSON() {
	storage := mem.NewMemStorage()
	svc := service.NewMetricsService(storage, nil, config.ServerConfig{})
	h := NewMetricsHandler(svc, nil)

	body := `{"id":"PollCount","type":"counter","delta":5}`
	req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	h.UpdateJSON(rec, req)

	fmt.Println(rec.Code)

	// Output:
	// 200
}

// ExampleMetricsHandler_UpdateBatch демонстрирует пакетное обновление метрик через JSON-массив.
func ExampleMetricsHandler_UpdateBatch() {
	storage := mem.NewMemStorage()
	svc := service.NewMetricsService(storage, nil, config.ServerConfig{})
	h := NewMetricsHandler(svc, nil)

	body := `[
		{"id":"Alloc","type":"gauge","value":100.5},
		{"id":"PollCount","type":"counter","delta":3}
	]`
	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	h.UpdateBatch(rec, req)

	fmt.Println(rec.Code)

	// Output:
	// 200
}

// ExampleMetricsHandler_Value демонстрирует получение значения gauge-метрики через URL-параметры.
func ExampleMetricsHandler_Value() {
	storage := mem.NewMemStorage()
	_ = storage.UpdateGauge(context.Background(), "Alloc", 42.0)
	svc := service.NewMetricsService(storage, nil, config.ServerConfig{})
	h := NewMetricsHandler(svc, nil)

	req := httptest.NewRequest(http.MethodGet, "/value/gauge/Alloc", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("type", "gauge")
	rctx.URLParams.Add("name", "Alloc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()
	h.Value(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(rec.Body.String())

	// Output:
	// 200
	// 42
}

// ExampleMetricsHandler_ValueJSON демонстрирует получение значения counter-метрики через JSON.
func ExampleMetricsHandler_ValueJSON() {
	storage := mem.NewMemStorage()
	_ = storage.UpdateCounter(context.Background(), "PollCount", 7)
	svc := service.NewMetricsService(storage, nil, config.ServerConfig{})
	h := NewMetricsHandler(svc, nil)

	body := `{"id":"PollCount","type":"counter"}`
	req := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	h.ValueJSON(rec, req)

	fmt.Println(rec.Code)

	// Output:
	// 200
}

// ExampleMetricsHandler_Ping демонстрирует проверку доступности хранилища.
func ExampleMetricsHandler_Ping() {
	storage := mem.NewMemStorage()
	svc := service.NewMetricsService(storage, nil, config.ServerConfig{})
	h := NewMetricsHandler(svc, nil)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	h.Ping(rec, req)

	fmt.Println(rec.Code)

	// Output:
	// 200
}

// ExampleMetricsHandler_GetList демонстрирует получение HTML-страницы со списком всех метрик.
func ExampleMetricsHandler_GetList() {
	storage := mem.NewMemStorage()
	svc := service.NewMetricsService(storage, nil, config.ServerConfig{})
	h := NewMetricsHandler(svc, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.GetList(rec, req)

	fmt.Println(rec.Code)

	// Output:
	// 200
}
