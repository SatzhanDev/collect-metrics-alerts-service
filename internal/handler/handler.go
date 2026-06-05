package handler

import (
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/audit"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/service"
)

type MetricsHandler struct {
	svc   service.Service
	audit *audit.Publisher
}

func NewMetricsHandler(svc service.Service, auditPublisher *audit.Publisher) *MetricsHandler {
	return &MetricsHandler{
		svc:   svc,
		audit: auditPublisher,
	}
}
