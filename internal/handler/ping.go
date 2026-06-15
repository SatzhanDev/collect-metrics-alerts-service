package handler

import (
	"context"
	"net/http"
	"time"
)

// Ping проверяет доступность хранилища и возвращает 200 OK при успехе GET /ping.
func (h *MetricsHandler) Ping(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	if err := h.svc.Ping(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
