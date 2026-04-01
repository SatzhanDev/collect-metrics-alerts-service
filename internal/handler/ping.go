package handler

import (
	"net/http"
)

func (h *MetricsHandler) Ping(w http.ResponseWriter, r *http.Request) {

	if h.db == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := h.db.PingContext(r.Context()); err != nil {
		http.Error(w, "database is unavailable", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
