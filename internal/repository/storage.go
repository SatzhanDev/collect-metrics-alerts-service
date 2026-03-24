package repository

import models "github.com/SatzhanDev/collect-metrics-alerts-service/internal/model"

type Storage interface {
	UpdateGauge(name string, value float64) error
	UpdateCounter(name string, delta int64) error
	GetGauge(name string) (float64, error)
	GetCounter(name string) (int64, error)
	GetAll() (gauges map[string]float64, counters map[string]int64)
	SetAll(gauges map[string]float64, counters map[string]int64) error
}

type FileStorage interface {
	SaveToFile(path string, metrics []models.Metrics) error
	RestoreFromFile(path string) ([]models.Metrics, error)
}
