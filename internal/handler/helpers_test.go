package handler

import (
	"net/http"
	"testing"

	models "github.com/SatzhanDev/collect-metrics-alerts-service/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestCollectMetricNames_MultipleMetrics(t *testing.T) {
	v := 1.0
	metrics := []models.Metrics{
		{ID: "cpu", MType: models.Gauge, Value: &v},
		{ID: "mem", MType: models.Gauge, Value: &v},
		{ID: "PollCount", MType: models.Counter},
	}
	names := collectMetricNames(metrics)
	assert.Equal(t, []string{"cpu", "mem", "PollCount"}, names)
}

func TestCollectMetricNames_EmptySlice(t *testing.T) {
	names := collectMetricNames([]models.Metrics{})
	assert.Empty(t, names)
}

func TestCollectMetricNames_NilSlice(t *testing.T) {
	names := collectMetricNames(nil)
	assert.Empty(t, names)
}

func TestClientIP_WithPort(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "192.168.1.100:54321"

	ip := clientIP(r)
	assert.Equal(t, "192.168.1.100", ip)
}

func TestClientIP_IPv6WithPort(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "[::1]:8080"

	ip := clientIP(r)
	assert.Equal(t, "::1", ip)
}

func TestClientIP_InvalidAddr(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "invalid-no-port"

	ip := clientIP(r)
	// При ошибке парсинга возвращается RemoteAddr как есть
	assert.Equal(t, "invalid-no-port", ip)
}
