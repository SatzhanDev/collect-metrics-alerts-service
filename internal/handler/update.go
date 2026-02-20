package handler

import (
	"net/http"

	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/service"
)

type MetricsHandler struct {
	svc service.MetricsService
}

func NewMetricsHandler(svc service.MetricsService) *MetricsHandler {
	return &MetricsHandler{
		svc: svc,
	}
}

func (h *MetricsHandler) Update(w http.ResponseWriter, r *http.Request) {

}
