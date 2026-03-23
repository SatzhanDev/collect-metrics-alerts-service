package service

import (
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/config"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/repository"
)

type MetricsService struct {
	storage repository.Storage
	cfg     config.ServerConfig
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

func NewMetricsService(storage repository.Storage, cfg config.ServerConfig) Service {
	return &MetricsService{
		storage: storage,
		cfg:     cfg,
	}
}

func (s *MetricsService) UpdateGauge(name string, value float64) error {

	if err := s.storage.UpdateGauge(name, value); err != nil {
		return err
	}
	if s.cfg.StoreInterval == 0 {
		return s.storage.SaveToFile(s.cfg.FilePath)
	}
	return nil
}

func (s *MetricsService) UpdateCounter(name string, delta int64) error {
	if err := s.storage.UpdateCounter(name, delta); err != nil {
		return err
	}
	if s.cfg.StoreInterval == 0 {
		return s.storage.SaveToFile(s.cfg.FilePath)
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
	return s.storage.RestoreFromFile(s.cfg.FilePath)
}
func (s *MetricsService) SaveToFile() error {
	return s.storage.SaveToFile(s.cfg.FilePath)
}
