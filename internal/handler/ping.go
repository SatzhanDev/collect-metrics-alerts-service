package handler

import (
	"context"
	"net/http"
	"time"
)

func (h *MetricsHandler) Ping(w http.ResponseWriter, r *http.Request) {

	if h.db == nil {
		http.Error(w, "database is not configured", http.StatusInternalServerError)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		http.Error(w, "database is unavailable", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
