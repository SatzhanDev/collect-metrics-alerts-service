package handler

import (
	"net/http"
	"strconv"
	"strings"

	models "github.com/SatzhanDev/collect-metrics-alerts-service/internal/model"
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

func (h *MetricsHandler) Update(w http.ResponseWriter, r *http.Request) {
	// log.Println("HANDLER HIT:", r.Method, r.URL.Path)
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")
	if parts[0] != "update" || len(parts) < 4 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	mType := parts[1]
	name := parts[2]
	valueStr := parts[3]

	if name == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if mType != models.Counter && mType != models.Gauge {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if mType == models.Gauge {
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.svc.UpdateGauge(name, value)
	}
	if mType == models.Counter {
		delta, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.svc.UpdateCounter(name, delta)
	}
	w.WriteHeader(http.StatusOK)

}
