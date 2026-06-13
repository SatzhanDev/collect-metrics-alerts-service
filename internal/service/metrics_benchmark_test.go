package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/config"
	models "github.com/SatzhanDev/collect-metrics-alerts-service/internal/models"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/repository/mem"
)

func BenchmarkMetricsService_UpdateBatch(b *testing.B) {

	storage := mem.NewMemStorage()
	cfg := config.ServerConfig{
		StoreInterval: 1,
	}
	svc := NewMetricsService(storage, nil, cfg)
	metrics := makeServiceBenchmarkMetrics(1000)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := svc.UpdateBatch(context.Background(), metrics); err != nil {
			b.Fatal(err)
		}
	}

}

func makeServiceBenchmarkMetrics(n int) []models.Metrics {

	metrics := make([]models.Metrics, 0, n)
	for i := 0; i < n/2; i++ {
		value := float64(i)
		metrics = append(metrics, models.Metrics{
			ID:    fmt.Sprintf("Gauge%d", i),
			MType: models.Gauge,
			Value: &value,
		})
	}
	for i := 0; i < n/2; i++ {
		delta := int64(i)
		metrics = append(metrics, models.Metrics{
			ID:    fmt.Sprintf("Counter%d", i),
			MType: models.Counter,
			Delta: &delta,
		})
	}
	return metrics

}
