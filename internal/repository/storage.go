// Package repository определяет интерфейсы хранилищ метрик.
package repository

import (
	"context"

	models "github.com/SatzhanDev/collect-metrics-alerts-service/internal/models"
)

// Storage описывает операции над хранилищем метрик в памяти или базе данных.
type Storage interface {
	// UpdateGauge сохраняет значение gauge-метрики с именем name.
	UpdateGauge(ctx context.Context, name string, value float64) error
	// UpdateCounter прибавляет delta к значению counter-метрики с именем name.
	UpdateCounter(ctx context.Context, name string, delta int64) error
	// GetGauge возвращает текущее значение gauge-метрики с именем name.
	GetGauge(ctx context.Context, name string) (float64, error)
	// GetCounter возвращает текущее значение counter-метрики с именем name.
	GetCounter(ctx context.Context, name string) (int64, error)
	// GetAll возвращает все gauge- и counter-метрики из хранилища.
	GetAll(ctx context.Context) (map[string]float64, map[string]int64, error)
	// SetAll заменяет все метрики в хранилище переданными значениями.
	SetAll(ctx context.Context, gauges map[string]float64, counters map[string]int64) error
	// UpdateBatch сохраняет набор метрик одним пакетным запросом.
	UpdateBatch(ctx context.Context, metrics []models.Metrics) error
	// Ping проверяет доступность хранилища.
	Ping(ctx context.Context) error
}

// FileStorage описывает операции сохранения и восстановления метрик из файла.
type FileStorage interface {
	// RestoreFromFile читает метрики из файла по указанному пути.
	RestoreFromFile(ctx context.Context, path string) ([]models.Metrics, error)
	// SaveToFile записывает метрики в файл по указанному пути.
	SaveToFile(ctx context.Context, path string, metrics []models.Metrics) error
}
