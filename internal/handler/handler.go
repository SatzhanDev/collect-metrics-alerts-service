package handler

import (
	"database/sql"

	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/service"
)

type MetricsHandler struct {
	svc service.Service
	db  *sql.DB
}

func NewMetricsHandler(svc service.Service, db *sql.DB) *MetricsHandler {
	return &MetricsHandler{
		svc: svc,
		db:  db,
	}
}
