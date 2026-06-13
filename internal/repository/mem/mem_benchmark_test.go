package mem

import (
	"context"
	"fmt"
	"testing"

	models "github.com/SatzhanDev/collect-metrics-alerts-service/internal/models"
)

func BenchmarkMemStorage_UpdateBatch(b *testing.B) {
	storage := NewMemStorage()
	metrics := makeBenchmarkMetrics(1000)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if err := storage.UpdateBatch(context.Background(), metrics); err != nil {
			b.Fatal(err)
		}
	}
}

func makeBenchmarkMetrics(n int) []models.Metrics {
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
