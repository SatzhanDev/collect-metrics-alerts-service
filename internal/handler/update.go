package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/logger"
	models "github.com/SatzhanDev/collect-metrics-alerts-service/internal/model"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

func (h *MetricsHandler) Update(w http.ResponseWriter, r *http.Request) {

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
		if err := h.svc.UpdateGauge(r.Context(), name, value); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	case models.Counter:
		delta, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid value", http.StatusBadRequest)
			return
		}
		if err := h.svc.UpdateCounter(r.Context(), name, delta); err != nil {
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
func (h *MetricsHandler) UpdateJSON(w http.ResponseWriter, r *http.Request) {

	var req, resp models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.ID == "" {
		logger.Log.Debug("id cannot be empty")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.MType == models.Gauge && req.Value == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	switch req.MType {
	case models.Gauge:
		if req.Value == nil {
			logger.Log.Debug("value cannot be nil")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := h.svc.UpdateGauge(r.Context(), req.ID, *req.Value); err != nil {
			logger.Log.Debug("cannot update gauge", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp.MType = req.MType
		resp.ID = req.ID
		resp.Value = req.Value
	case models.Counter:
		if req.Delta == nil {
			logger.Log.Debug("delta cannot be nil")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := h.svc.UpdateCounter(r.Context(), req.ID, *req.Delta); err != nil {
			logger.Log.Debug("cannot update counter", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp.MType = req.MType
		resp.ID = req.ID
		resp.Delta = req.Delta
	default:
		logger.Log.Debug("invalid metric type or empty metric type", zap.String("type", req.MType))
		w.WriteHeader(http.StatusBadRequest)
		return
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
