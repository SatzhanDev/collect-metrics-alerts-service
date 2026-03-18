package main

import (
	"errors"
	"log"
	"net"
	"net/http"

	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/handler"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/logger"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/repository"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/service"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

func main() {
	cfg := parseFlags()
	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatal(err)
	}
	defer logger.Log.Sync()

	storage := repository.NewMemStorage()
	svc := service.NewMetricsService(storage)
	h := handler.NewMetricsHandler(svc)

	r := chi.NewRouter()
	r.Get("/", h.GetList)
	r.Get("/value/{type}/{name}", h.Value)
	r.Post("/update/{type}/{name}/{value}", h.Update)

	r.Post("/update", h.UpdateJSON)
	r.Post("/value", h.ValueJSON)

	addr := normalizeAddr(cfg.Addr)

	logger.Log.Info("Running server", zap.String("address", cfg.Addr))
	if err := http.ListenAndServe(addr, logger.WithLogging(r)); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
