package main

import (
	"log"
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

	log.Printf("SERVER STARTED on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, r); err != nil {
		log.Fatal(err)
	}

}
