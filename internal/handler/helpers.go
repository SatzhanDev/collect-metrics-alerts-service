package handler

import (
	"net"
	"net/http"

	models "github.com/SatzhanDev/collect-metrics-alerts-service/internal/models"
)

func collectMetricNames(metrics []models.Metrics) []string {
	names := make([]string, 0, len(metrics))

	for _, metric := range metrics {
		names = append(names, metric.ID)
	}
	return names
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
