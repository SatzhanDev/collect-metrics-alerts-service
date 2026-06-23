// Package file реализует хранилище метрик на основе JSON-файла.
package file

import (
	"context"
	"encoding/json"
	"os"

	models "github.com/SatzhanDev/collect-metrics-alerts-service/internal/models"
)

type JSONFileStorage struct{}

func NewJSONFileStorage() *JSONFileStorage {
	return &JSONFileStorage{}
}

func (s *JSONFileStorage) RestoreFromFile(ctx context.Context, path string) ([]models.Metrics, error) {

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var metrics []models.Metrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		return nil, err
	}
	return metrics, nil

}

func (s *JSONFileStorage) SaveToFile(ctx context.Context, path string, metrics []models.Metrics) error {

	data, err := json.Marshal(metrics)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
