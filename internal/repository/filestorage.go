package repository

import (
	"encoding/json"
	"os"

	models "github.com/SatzhanDev/collect-metrics-alerts-service/internal/model"
)

type JSONFileStorage struct{}

func NewJSONFileStorage() *JSONFileStorage {
	return &JSONFileStorage{}
}

func (s *JSONFileStorage) RestoreFromFile(path string) ([]models.Metrics, error) {

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

func (s *JSONFileStorage) SaveToFile(path string, metrics []models.Metrics) error {

	data, err := json.Marshal(metrics)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
