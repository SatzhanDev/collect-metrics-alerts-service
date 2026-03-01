package handler

import (
	"net/http"
	"strconv"

	models "github.com/SatzhanDev/collect-metrics-alerts-service/internal/model"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/service"
	"github.com/go-chi/chi"
)

type MetricsHandler struct {
	svc service.Service
}

func NewMetricsHandler(svc service.Service) *MetricsHandler {
	return &MetricsHandler{
		svc: svc,
	}
}

func (h *MetricsHandler) Update(w http.ResponseWriter, r *http.Request) {
	// log.Println("RECEIVED:", r.Method, r.URL.Path)
	mType := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")
	valueStr := chi.URLParam(r, "value")

	switch mType {
	case models.Gauge:
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			http.Error(w, "invalid value", http.StatusBadRequest)
			return
		}
		if err := h.svc.UpdateGauge(name, value); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	case models.Counter:
		delta, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid value", http.StatusBadRequest)
			return
		}
		if err := h.svc.UpdateCounter(name, delta); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "invalid metric type", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

}
