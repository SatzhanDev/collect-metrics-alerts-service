package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/config"
	models "github.com/SatzhanDev/collect-metrics-alerts-service/internal/models"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/repository/file"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/repository/mem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// errStorage — хранилище, всегда возвращающее ошибку (для тестирования error-веток).
type errStorage struct{}

func (e errStorage) UpdateGauge(_ context.Context, _ string, _ float64) error {
	return errors.New("storage error")
}
func (e errStorage) UpdateCounter(_ context.Context, _ string, _ int64) error {
	return errors.New("storage error")
}
func (e errStorage) GetGauge(_ context.Context, _ string) (float64, error) {
	return 0, errors.New("storage error")
}
func (e errStorage) GetCounter(_ context.Context, _ string) (int64, error) {
	return 0, errors.New("storage error")
}
func (e errStorage) GetAll(_ context.Context) (map[string]float64, map[string]int64, error) {
	return nil, nil, errors.New("storage error")
}
func (e errStorage) SetAll(_ context.Context, _ map[string]float64, _ map[string]int64) error {
	return errors.New("storage error")
}
func (e errStorage) UpdateBatch(_ context.Context, _ []models.Metrics) error {
	return errors.New("storage error")
}
func (e errStorage) Ping(_ context.Context) error {
	return errors.New("storage error")
}

// errFileStorage — файловое хранилище, всегда возвращающее ошибку.
type errFileStorage struct{}

func (e errFileStorage) RestoreFromFile(_ context.Context, _ string) ([]models.Metrics, error) {
	return nil, errors.New("file storage error")
}
func (e errFileStorage) SaveToFile(_ context.Context, _ string, _ []models.Metrics) error {
	return errors.New("file storage error")
}

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

func TestMetricsService_UpdateCounter(t *testing.T) {
	storage := mem.NewMemStorage()

	svc := NewMetricsService(
		storage,
		nil,
		config.ServerConfig{},
	)

	err := svc.UpdateCounter(t.Context(), "PollCount", 10)
	require.NoError(t, err)

	err = svc.UpdateCounter(t.Context(), "PollCount", 5)
	require.NoError(t, err)

	value, err := svc.GetCounter(t.Context(), "PollCount")
	require.NoError(t, err)

	assert.Equal(t, int64(15), value)
}

func TestMetricsService_UpdateCounter_ImmediateSave(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "metrics.json")

	storage := mem.NewMemStorage()
	fileStorage := file.NewJSONFileStorage()

	svc := NewMetricsService(storage, fileStorage, config.ServerConfig{
		FileStoragePath: tmpFile,
		StoreInterval:   0,
	})

	err := svc.UpdateCounter(t.Context(), "PollCount", 42)
	require.NoError(t, err)

	data, err := os.ReadFile(tmpFile)
	require.NoError(t, err)

	assert.Contains(t, string(data), `"PollCount"`)
	assert.Contains(t, string(data), `"counter"`)
}

func TestMetricsService_GetGauge(t *testing.T) {
	storage := mem.NewMemStorage()

	err := storage.UpdateGauge(t.Context(), "Alloc", 123.45)
	require.NoError(t, err)

	svc := NewMetricsService(
		storage,
		nil,
		config.ServerConfig{},
	)

	value, err := svc.GetGauge(t.Context(), "Alloc")
	require.NoError(t, err)

	assert.Equal(t, 123.45, value)
}

func TestMetricsService_GetCounter(t *testing.T) {
	storage := mem.NewMemStorage()

	err := storage.UpdateCounter(t.Context(), "PollCount", 77)
	require.NoError(t, err)

	svc := NewMetricsService(
		storage,
		nil,
		config.ServerConfig{},
	)

	value, err := svc.GetCounter(t.Context(), "PollCount")
	require.NoError(t, err)

	assert.Equal(t, int64(77), value)
}

func TestMetricsService_GetAll(t *testing.T) {
	storage := mem.NewMemStorage()

	require.NoError(t, storage.UpdateGauge(t.Context(), "Alloc", 1.5))
	require.NoError(t, storage.UpdateCounter(t.Context(), "PollCount", 3))

	svc := NewMetricsService(
		storage,
		nil,
		config.ServerConfig{},
	)

	gauges, counters, err := svc.GetAll(t.Context())

	require.NoError(t, err)

	assert.Equal(t, 1.5, gauges["Alloc"])
	assert.Equal(t, int64(3), counters["PollCount"])
}

func TestMetricsService_Ping(t *testing.T) {
	storage := mem.NewMemStorage()

	svc := NewMetricsService(
		storage,
		nil,
		config.ServerConfig{},
	)

	err := svc.Ping(t.Context())
	require.NoError(t, err)
}

func TestMetricsService_UpdateBatch(t *testing.T) {
	storage := mem.NewMemStorage()

	svc := NewMetricsService(
		storage,
		nil,
		config.ServerConfig{},
	)

	value := 100.5
	delta := int64(10)

	metrics := []models.Metrics{
		{
			ID:    "Alloc",
			MType: models.Gauge,
			Value: &value,
		},
		{
			ID:    "PollCount",
			MType: models.Counter,
			Delta: &delta,
		},
	}

	err := svc.UpdateBatch(t.Context(), metrics)
	require.NoError(t, err)

	gauge, err := storage.GetGauge(t.Context(), "Alloc")
	require.NoError(t, err)
	assert.Equal(t, value, gauge)

	counter, err := storage.GetCounter(t.Context(), "PollCount")
	require.NoError(t, err)
	assert.Equal(t, delta, counter)
}
func TestMetricsService_RestoreFromFile(t *testing.T) {
	tmpFile := filepath.Join(
		t.TempDir(),
		"metrics.json",
	)

	fileStorage := file.NewJSONFileStorage()
	memStorage := mem.NewMemStorage()

	value := 123.45
	delta := int64(7)

	err := fileStorage.SaveToFile(
		t.Context(),
		tmpFile,
		[]models.Metrics{
			{
				ID:    "Alloc",
				MType: models.Gauge,
				Value: &value,
			},
			{
				ID:    "PollCount",
				MType: models.Counter,
				Delta: &delta,
			},
		},
	)
	require.NoError(t, err)

	svc := NewMetricsService(
		memStorage,
		fileStorage,
		config.ServerConfig{
			FileStoragePath: tmpFile,
		},
	)

	err = svc.RestoreFromFile(t.Context())
	require.NoError(t, err)

	gauge, err := memStorage.GetGauge(
		t.Context(),
		"Alloc",
	)
	require.NoError(t, err)
	require.Equal(t, value, gauge)

	counter, err := memStorage.GetCounter(
		t.Context(),
		"PollCount",
	)
	require.NoError(t, err)
	require.Equal(t, delta, counter)
}

// ─── Error branch tests using errStorage ──────────────────────────────────

func TestMetricsService_UpdateGauge_StorageError(t *testing.T) {
	svc := NewMetricsService(errStorage{}, nil, config.ServerConfig{})
	err := svc.UpdateGauge(t.Context(), "cpu", 1.0)
	require.Error(t, err)
}

func TestMetricsService_UpdateCounter_StorageError(t *testing.T) {
	svc := NewMetricsService(errStorage{}, nil, config.ServerConfig{})
	err := svc.UpdateCounter(t.Context(), "PollCount", 5)
	require.Error(t, err)
}

func TestMetricsService_UpdateCounter_NoImmediateSave(t *testing.T) {
	// StoreInterval > 0: сохранять сразу НЕ надо → return nil
	storage := mem.NewMemStorage()
	svc := NewMetricsService(storage, nil, config.ServerConfig{
		StoreInterval: 10 * time.Second,
	})
	err := svc.UpdateCounter(t.Context(), "PollCount", 3)
	require.NoError(t, err)
}

func TestMetricsService_GetGauge_StorageError(t *testing.T) {
	svc := NewMetricsService(errStorage{}, nil, config.ServerConfig{})
	_, err := svc.GetGauge(t.Context(), "cpu")
	require.Error(t, err)
}

func TestMetricsService_GetCounter_StorageError(t *testing.T) {
	svc := NewMetricsService(errStorage{}, nil, config.ServerConfig{})
	_, err := svc.GetCounter(t.Context(), "PollCount")
	require.Error(t, err)
}

func TestMetricsService_GetAll_StorageError(t *testing.T) {
	svc := NewMetricsService(errStorage{}, nil, config.ServerConfig{})
	_, _, err := svc.GetAll(t.Context())
	require.Error(t, err)
}

func TestMetricsService_RestoreFromFile_FileStorageError(t *testing.T) {
	svc := NewMetricsService(mem.NewMemStorage(), errFileStorage{}, config.ServerConfig{
		FileStoragePath: "/some/path",
	})
	err := svc.RestoreFromFile(t.Context())
	require.Error(t, err)
}

func TestMetricsService_SaveToFile_StorageError(t *testing.T) {
	// GetAll возвращает ошибку → SaveToFile должен её вернуть
	svc := NewMetricsService(errStorage{}, file.NewJSONFileStorage(), config.ServerConfig{
		FileStoragePath: "/tmp/metrics_test.json",
	})
	err := svc.SaveToFile(t.Context())
	require.Error(t, err)
}
