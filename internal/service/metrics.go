package service

import (
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/config"
	models "github.com/SatzhanDev/collect-metrics-alerts-service/internal/model"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/repository"
)

type MetricsService struct {
	storage     repository.Storage
	filestorage repository.FileStorage
	cfg         config.ServerConfig
}
type Service interface {
	UpdateGauge(name string, value float64) error
	UpdateCounter(name string, delta int64) error
	GetGauge(name string) (float64, error)
	GetCounter(name string) (int64, error)
	GetAll() (gauges map[string]float64, counters map[string]int64)
	RestoreFromFile() error
	SaveToFile() error
}

func NewMetricsService(storage repository.Storage, filestorage repository.FileStorage, cfg config.ServerConfig) Service {
	return &MetricsService{
		storage:     storage,
		filestorage: filestorage,
		cfg:         cfg,
	}
}

func (s *MetricsService) UpdateGauge(name string, value float64) error {

	if err := s.storage.UpdateGauge(name, value); err != nil {
		return err
	}
	if s.cfg.StoreInterval == 0 {
		return s.SaveToFile()
	}
	return nil
}

func (s *MetricsService) UpdateCounter(name string, delta int64) error {
	if err := s.storage.UpdateCounter(name, delta); err != nil {
		return err
	}
	if s.cfg.StoreInterval == 0 {
		return s.SaveToFile()
	}
	return nil
}
func (s *MetricsService) GetGauge(name string) (float64, error) {
	res, err := s.storage.GetGauge(name)
	if err != nil {
		return 0, err
	}
	return res, nil
}
func (s *MetricsService) GetCounter(name string) (int64, error) {
	res, err := s.storage.GetCounter(name)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (s *MetricsService) GetAll() (gauges map[string]float64, counters map[string]int64) {
	return s.storage.GetAll()
}
func (s *MetricsService) RestoreFromFile() error {
	metrics, err := s.filestorage.RestoreFromFile(s.cfg.FilePath)
	if err != nil {
		return err
	}

	gauges := make(map[string]float64)
	counters := make(map[string]int64)

	for _, metric := range metrics {
		switch metric.MType {
		case models.Gauge:
			if metric.Value != nil {
				gauges[metric.ID] = *metric.Value
			}
		case models.Counter:
			if metric.Delta != nil {
				counters[metric.ID] = *metric.Delta
			}
		}
	}

	return s.storage.SetAll(gauges, counters)
}
func (s *MetricsService) SaveToFile() error {
	gauges, counters := s.storage.GetAll()

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

	return s.filestorage.SaveToFile(s.cfg.FilePath, metrics)
}
