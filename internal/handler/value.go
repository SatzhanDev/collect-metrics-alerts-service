package handler

import (
	"io"
	"net/http"
	"strconv"

	models "github.com/SatzhanDev/collect-metrics-alerts-service/internal/model"
	"github.com/go-chi/chi"
)

func (h *MetricsHandler) Value(w http.ResponseWriter, r *http.Request) {

	mType := chi.URLParam(r, "type")
	if mType != models.Counter && mType != models.Gauge {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	mName := chi.URLParam(r, "name")
	if mName == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var res string
	if mType == models.Gauge {
		value, err := h.svc.GetGauge(mName)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		res = strconv.FormatFloat(value, 'f', -1, 64)

	}
	if mType == models.Counter {
		value, err := h.svc.GetCounter(mName)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		res = strconv.FormatInt(value, 10)
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, res)

}
