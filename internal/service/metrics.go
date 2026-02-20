package service

import "github.com/SatzhanDev/collect-metrics-alerts-service/internal/repository"

type MetricsService struct {
	storage repository.Storage
}
type Service interface {
	UpdateGauge(name string, value float64) error
	UpdateCounter(name string, delta int64) error
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
