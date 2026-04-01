package repository

import (
	"context"

	models "github.com/SatzhanDev/collect-metrics-alerts-service/internal/model"
)

type Storage interface {
	UpdateGauge(ctx context.Context, name string, value float64) error
	UpdateCounter(ctx context.Context, name string, delta int64) error
	GetGauge(ctx context.Context, name string) (float64, error)
	GetCounter(ctx context.Context, name string) (int64, error)
	GetAll(ctx context.Context) (map[string]float64, map[string]int64, error)
	SetAll(ctx context.Context, gauges map[string]float64, counters map[string]int64) error
}

type FileStorage interface {
	RestoreFromFile(ctx context.Context, path string) ([]models.Metrics, error)
	SaveToFile(ctx context.Context, path string, metrics []models.Metrics) error
}
