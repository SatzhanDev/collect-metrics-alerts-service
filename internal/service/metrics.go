package service

import "github.com/SatzhanDev/collect-metrics-alerts-service/internal/repository"

type MetricsService struct {
	storage repository.Storage
}
type Service interface {
	UpdateGauge(name string, value float64) error
	UpdateCounter(name string, delta int64) error
	GetGauge(name string) (float64, error)
	GetCounter(name string) (int64, error)
	Snapshot() (gauges map[string]float64, counters map[string]int64)
}

func NewMetricsService(storage repository.Storage) Service {
	return &MetricsService{
		storage: storage,
	}
}

func (s *MetricsService) UpdateGauge(name string, value float64) error {
	if err := s.storage.UpdateGauge(name, value); err != nil {
		return err
	}
	return nil
}

func (s *MetricsService) UpdateCounter(name string, delta int64) error {
	if err := s.storage.UpdateCounter(name, delta); err != nil {
		return err
	}
	return nil
}
func (s *MetricsService) GetGauge(name string) (float64, error) {
	res, err := s.GetGauge(name)
	if err != nil {
		return 0, err
	}
	return res, nil
}
func (s *MetricsService) GetCounter(name string) (int64, error) {
	res, err := s.GetCounter(name)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (s *MetricsService) Snapshot() (gauges map[string]float64, counters map[string]int64) {
	return nil, nil
}
