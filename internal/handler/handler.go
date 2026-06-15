// Package handler содержит HTTP-обработчики для работы с метриками.
package handler

import (
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/audit"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/service"
)

// MetricsHandler обрабатывает HTTP-запросы для получения и обновления метрик.
type MetricsHandler struct {
	svc   service.Service
	audit *audit.Publisher
}

// NewMetricsHandler создаёт новый MetricsHandler с указанным сервисом и издателем аудита.
func NewMetricsHandler(svc service.Service, auditPublisher *audit.Publisher) *MetricsHandler {
	return &MetricsHandler{
		svc:   svc,
		audit: auditPublisher,
	}
}
