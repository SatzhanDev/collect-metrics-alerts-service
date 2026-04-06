package handler

import (
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/service"
)

type MetricsHandler struct {
	svc service.Service
}

func NewMetricsHandler(svc service.Service) *MetricsHandler {
	return &MetricsHandler{
		svc: svc,
	}
}
