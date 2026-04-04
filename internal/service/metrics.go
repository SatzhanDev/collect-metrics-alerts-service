package service

import (
	"context"

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
	UpdateGauge(ctx context.Context, name string, value float64) error
	UpdateCounter(ctx context.Context, name string, delta int64) error
	GetGauge(ctx context.Context, name string) (float64, error)
	GetCounter(ctx context.Context, name string) (int64, error)
	GetAll(ctx context.Context) (map[string]float64, map[string]int64, error)
	RestoreFromFile(ctx context.Context) error
	SaveToFile(ctx context.Context) error
	UpdateBatch(ctx context.Context, metrics []models.Metrics) error
}

func NewMetricsService(storage repository.Storage, filestorage repository.FileStorage, cfg config.ServerConfig) Service {
	return &MetricsService{
		storage:     storage,
		filestorage: filestorage,
		cfg:         cfg,
	}
}

func (s *MetricsService) UpdateGauge(ctx context.Context, name string, value float64) error {

	if err := s.storage.UpdateGauge(ctx, name, value); err != nil {
		return err
	}
	if s.cfg.StoreInterval == 0 {
		return s.SaveToFile(ctx)
	}
	return nil
}

func (s *MetricsService) UpdateCounter(ctx context.Context, name string, delta int64) error {
	if err := s.storage.UpdateCounter(ctx, name, delta); err != nil {
		return err
	}
	if s.cfg.StoreInterval == 0 {
		return s.SaveToFile(ctx)
	}
	return nil
}
func (s *MetricsService) GetGauge(ctx context.Context, name string) (float64, error) {
	res, err := s.storage.GetGauge(ctx, name)
	if err != nil {
		return 0, err
	}
	return res, nil
}
func (s *MetricsService) GetCounter(ctx context.Context, name string) (int64, error) {
	res, err := s.storage.GetCounter(ctx, name)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (s *MetricsService) GetAll(ctx context.Context) (map[string]float64, map[string]int64, error) {
	gauges, counters, err := s.storage.GetAll(ctx)
	if err != nil {
		return nil, nil, err
	}
	return gauges, counters, nil
}
func (s *MetricsService) RestoreFromFile(ctx context.Context) error {
	metrics, err := s.filestorage.RestoreFromFile(ctx, s.cfg.FileStoragePath)
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
	return s.storage.SetAll(ctx, gauges, counters)
}
func (s *MetricsService) SaveToFile(ctx context.Context) error {
	if s.filestorage == nil {
		return nil
	}
	gauges, counters, err := s.storage.GetAll(ctx)
	if err != nil {
		return err
	}

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

	return s.filestorage.SaveToFile(ctx, s.cfg.FileStoragePath, metrics)
}

func (s *MetricsService) UpdateBatch(ctx context.Context, metrics []models.Metrics) error {
	return s.storage.UpdateBatch(ctx, metrics)
}
