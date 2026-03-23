package repository

import (
	"encoding/json"
	"os"

	models "github.com/SatzhanDev/collect-metrics-alerts-service/internal/model"
)

func (s *MemStorage) RestoreFromFile(path string) error {

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var metrics []models.Metrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, metric := range metrics {
		switch metric.MType {
		case models.Gauge:
			if metric.Value != nil {
				s.gauges[metric.ID] = *metric.Value
			}
		case models.Counter:
			if metric.Delta != nil {
				s.counters[metric.ID] = *metric.Delta
			}
		}
	}
	return nil
}

func (s *MemStorage) SaveToFile(path string) error {
	gauges, counters := s.GetAll()
	var metrics []models.Metrics

	for id, value := range gauges {
		v := value
		metrics = append(metrics, models.Metrics{
			ID:    id,
			MType: models.Gauge,
			Value: &v,
		})
	}
	for id, delta := range counters {
		d := delta
		metrics = append(metrics, models.Metrics{
			ID:    id,
			MType: models.Counter,
			Delta: &d,
		})
	}
	data, err := json.Marshal(metrics)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
