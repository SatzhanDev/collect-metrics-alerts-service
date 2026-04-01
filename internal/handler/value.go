package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/logger"
	models "github.com/SatzhanDev/collect-metrics-alerts-service/internal/model"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
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
		value, err := h.svc.GetGauge(r.Context(), mName)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		res = strconv.FormatFloat(value, 'f', -1, 64)

	}
	if mType == models.Counter {
		value, err := h.svc.GetCounter(r.Context(), mName)
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

func (h *MetricsHandler) ValueJSON(w http.ResponseWriter, r *http.Request) {

	var req, resp models.Metrics

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.MType != models.Counter && req.MType != models.Gauge {
		logger.Log.Debug("invalid metric type or empty metric type", zap.String("type", req.MType))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	resp.MType = req.MType

	if req.ID == "" {
		logger.Log.Debug("id cannot be empty")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	resp.ID = req.ID
	switch req.MType {
	case models.Gauge:
		value, err := h.svc.GetGauge(r.Context(), req.ID)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		resp.Value = &value

	case models.Counter:
		value, err := h.svc.GetCounter(r.Context(), req.ID)
		if err != nil {
			logger.Log.Debug("id not found", zap.String("type", req.MType))
			w.WriteHeader(http.StatusNotFound)
			return
		}
		resp.Delta = &value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Log.Debug("error encoding response", zap.Error(err))
		return
	}
	logger.Log.Debug("sending HTTP 200 response")

}
