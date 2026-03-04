package main

import (
	"errors"
	"log"
	"net"
	"net/http"

	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/handler"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/repository"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/service"
	"github.com/go-chi/chi"
)

func main() {
	cfg := parseFlags()

	storage := repository.NewMemStorage()
	svc := service.NewMetricsService(storage)
	h := handler.NewMetricsHandler(svc)

	r := chi.NewRouter()
	r.Get("/", h.GetList)
	r.Get("/value/{type}/{name}", h.Value)
	r.Post("/update/{type}/{name}/{value}", h.Update)

	addr := normalizeAddr(cfg.Addr)

	log.Printf("SERVER STARTED on %s", cfg.Addr)
	if err := http.ListenAndServe(addr, r); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}

}
func normalizeAddr(in string) string {
	host, port, err := net.SplitHostPort(in)
	if err != nil {
		return in
	}
	if host == "localhost" {
		return ":" + port
	}
	return in
}
