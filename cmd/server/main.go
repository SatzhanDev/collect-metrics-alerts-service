package main

import (
	"log"
	"net/http"

	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/handler"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/repository"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/service"
)

func main() {
	storage := repository.NewMemStorage()
	svc := service.NewMetricsService(storage)
	h := handler.NewMetricsHandler(svc)
	// mux := http.NewServeMux()
	// mux.HandleFunc("/", h.Update)
	log.Println("SERVER STARTED on :8080")
	if err := http.ListenAndServe(":8080", http.HandlerFunc(h.Update)); err != nil {
		log.Fatal(err)
	}

	// http.ListenAndServe(":8080", mux)
}
