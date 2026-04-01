package service

import (
	"os"
	"testing"
	"time"

	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/config"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/repository/file"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/repository/mem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsService_UpdateGauge_NoImmediateSave(t *testing.T) {
	storage := mem.NewMemStorage()
	cfg := config.ServerConfig{
		FileStoragePath: "test_metrics.json",
		StoreInterval:   10 * time.Second,
	}
	fileStorage := file.NewJSONFileStorage()

	svc := NewMetricsService(storage, fileStorage, cfg)

	err := svc.UpdateGauge(t.Context(), "cpu", 12.5)
	require.NoError(t, err)

	value, err := storage.GetGauge(t.Context(), "cpu")
	require.NoError(t, err)
	assert.Equal(t, 12.5, value)
}

func TestMetricsService_UpdateGauge_ImmediateSave(t *testing.T) {
	tmpFile := "test_metrics_save.json"
	defer os.Remove(tmpFile)

	storage := mem.NewMemStorage()
	fileStorage := file.NewJSONFileStorage()

	cfg := config.ServerConfig{
		FileStoragePath: tmpFile,
		StoreInterval:   0,
	}

	svc := NewMetricsService(storage, fileStorage, cfg)

	err := svc.UpdateGauge(t.Context(), "cpu", 99.9)
	require.NoError(t, err)

	data, err := os.ReadFile(tmpFile)
	require.NoError(t, err)

	assert.Contains(t, string(data), `"id":"cpu"`)
	assert.Contains(t, string(data), `"value":99.9`)
}
