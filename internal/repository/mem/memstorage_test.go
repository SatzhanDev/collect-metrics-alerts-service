package mem

import (
	"context"
	"testing"

	models "github.com/SatzhanDev/collect-metrics-alerts-service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemStorage_UpdateGauge(t *testing.T) {
	s := NewMemStorage()

	err := s.UpdateGauge(t.Context(), "cpu", 10.5)
	require.NoError(t, err)

	val, err := s.GetGauge(t.Context(), "cpu")
	require.NoError(t, err)
	assert.Equal(t, 10.5, val)
}

func TestMemStorage_UpdateCounter(t *testing.T) {

	ctx := context.Background()
	storage := NewMemStorage()
	err := storage.UpdateCounter(ctx, "PollCount", 10)
	require.NoError(t, err)
	err = storage.UpdateCounter(ctx, "PollCount", 5)
	require.NoError(t, err)
	got, err := storage.GetCounter(ctx, "PollCount")
	require.NoError(t, err)
	require.Equal(t, int64(15), got)

}

func TestMemStorage_GetCounterNotFound(t *testing.T) {

	ctx := context.Background()
	storage := NewMemStorage()
	_, err := storage.GetCounter(ctx, "unknown")
	require.Error(t, err)

}

func TestMemStorage_GetAll(t *testing.T) {

	ctx := context.Background()
	storage := NewMemStorage()
	require.NoError(t, storage.UpdateGauge(ctx, "Alloc", 12.5))
	require.NoError(t, storage.UpdateCounter(ctx, "PollCount", 3))
	gauges, counters, err := storage.GetAll(ctx)
	require.NoError(t, err)
	require.Equal(t, 12.5, gauges["Alloc"])
	require.Equal(t, int64(3), counters["PollCount"])

}

func TestMemStorage_SetAll(t *testing.T) {

	ctx := context.Background()
	storage := NewMemStorage()
	gauges := map[string]float64{
		"Alloc": 10.5,
	}
	counters := map[string]int64{
		"PollCount": 7,
	}
	err := storage.SetAll(ctx, gauges, counters)
	require.NoError(t, err)
	gotGauge, err := storage.GetGauge(ctx, "Alloc")
	require.NoError(t, err)
	require.Equal(t, 10.5, gotGauge)
	gotCounter, err := storage.GetCounter(ctx, "PollCount")
	require.NoError(t, err)
	require.Equal(t, int64(7), gotCounter)

}

func TestMemStorage_UpdateBatch(t *testing.T) {

	ctx := context.Background()
	storage := NewMemStorage()
	gaugeValue := 100.5
	counterDelta := int64(4)
	metrics := []models.Metrics{
		{
			ID:    "Alloc",
			MType: models.Gauge,
			Value: &gaugeValue,
		},
		{
			ID:    "PollCount",
			MType: models.Counter,
			Delta: &counterDelta,
		},
	}
	err := storage.UpdateBatch(ctx, metrics)
	require.NoError(t, err)
	gotGauge, err := storage.GetGauge(ctx, "Alloc")
	require.NoError(t, err)
	require.Equal(t, gaugeValue, gotGauge)
	gotCounter, err := storage.GetCounter(ctx, "PollCount")
	require.NoError(t, err)
	require.Equal(t, counterDelta, gotCounter)

}

func TestMemStorage_UpdateBatchErrors(t *testing.T) {

	ctx := context.Background()
	t.Run("empty id", func(t *testing.T) {
		storage := NewMemStorage()
		value := 1.1
		err := storage.UpdateBatch(ctx, []models.Metrics{
			{
				ID:    "",
				MType: models.Gauge,
				Value: &value,
			},
		})
		require.Error(t, err)
	})
	t.Run("nil gauge value", func(t *testing.T) {
		storage := NewMemStorage()
		err := storage.UpdateBatch(ctx, []models.Metrics{
			{
				ID:    "Alloc",
				MType: models.Gauge,
				Value: nil,
			},
		})
		require.Error(t, err)
	})
	t.Run("nil counter delta", func(t *testing.T) {
		storage := NewMemStorage()
		err := storage.UpdateBatch(ctx, []models.Metrics{
			{
				ID:    "PollCount",
				MType: models.Counter,
				Delta: nil,
			},
		})
		require.Error(t, err)
	})
	t.Run("invalid metric type", func(t *testing.T) {
		storage := NewMemStorage()
		err := storage.UpdateBatch(ctx, []models.Metrics{
			{
				ID:    "Unknown",
				MType: "unknown",
			},
		})
		require.Error(t, err)
	})

}

func TestMemStorage_Ping(t *testing.T) {

	ctx := context.Background()
	storage := NewMemStorage()
	err := storage.Ping(ctx)
	require.NoError(t, err)

}
