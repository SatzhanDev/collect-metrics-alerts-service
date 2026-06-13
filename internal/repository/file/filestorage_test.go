package file

import (
	"testing"

	models "github.com/SatzhanDev/collect-metrics-alerts-service/internal/models"
	"github.com/stretchr/testify/require"
)

func TestJSONFileStorage_SaveAndRestore(t *testing.T) {
	tmpFile := t.TempDir() + "/metrics.json"

	storage := NewJSONFileStorage()

	value := 123.45
	delta := int64(5)

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

	err := storage.SaveToFile(t.Context(), tmpFile, metrics)
	require.NoError(t, err)

	restored, err := storage.RestoreFromFile(t.Context(), tmpFile)
	require.NoError(t, err)

	require.Len(t, restored, 2)
}

func TestJSONFileStorage_RestoreMissingFile(t *testing.T) {
	storage := NewJSONFileStorage()

	metrics, err := storage.RestoreFromFile(
		t.Context(),
		"definitely-not-existing-file.json",
	)

	require.NoError(t, err)
	require.Nil(t, metrics)
}
