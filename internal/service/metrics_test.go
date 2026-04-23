package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/config"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/repository/file"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/repository/mem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsService_UpdateGauge_NoImmediateSave(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "metrics.json")

	storage := mem.NewMemStorage()
	fileStorage := file.NewJSONFileStorage()

	cfg := config.ServerConfig{
		FileStoragePath: tmpFile,
		StoreInterval:   10 * time.Second,
	}

	svc := NewMetricsService(storage, fileStorage, cfg)

	err := svc.UpdateGauge(t.Context(), "cpu", 12.5)
	require.NoError(t, err)

	value, err := storage.GetGauge(t.Context(), "cpu")
	require.NoError(t, err)
	assert.Equal(t, 12.5, value)

	_, err = os.Stat(tmpFile)
	assert.True(t, os.IsNotExist(err), "file should not be created when StoreInterval > 0")
}

func TestMetricsService_UpdateGauge_ImmediateSave(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "metrics.json")

	storage := mem.NewMemStorage()
	fileStorage := file.NewJSONFileStorage()

	cfg := config.ServerConfig{
		FileStoragePath: tmpFile,
		StoreInterval:   0,
	}

	svc := NewMetricsService(storage, fileStorage, cfg)

	err := svc.UpdateGauge(t.Context(), "cpu", 99.9)
	require.NoError(t, err)

	value, err := storage.GetGauge(t.Context(), "cpu")
	require.NoError(t, err)
	assert.Equal(t, 99.9, value)

	data, err := os.ReadFile(tmpFile)
	require.NoError(t, err)

	assert.Contains(t, string(data), `"id":"cpu"`)
	assert.Contains(t, string(data), `"value":99.9`)
}
