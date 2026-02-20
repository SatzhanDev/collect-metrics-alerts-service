package service

import "github.com/SatzhanDev/collect-metrics-alerts-service/internal/repository"

type MetricsService struct {
	storage repository.Storage
}

func NewMetricsService(storage repository.Storage) *MetricsService {
	return &MetricsService{
		storage: storage,
	}
}

func UpdateGauge(name string, value float64) error {
	return nil
}

func UpdateCounter(name string, delta int64) error {
	return nil
}
