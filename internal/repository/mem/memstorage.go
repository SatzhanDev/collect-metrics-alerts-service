package mem

import (
	"context"
	"errors"
	"sync"

	models "github.com/SatzhanDev/collect-metrics-alerts-service/internal/model"
)

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
	mu       sync.RWMutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (s *MemStorage) UpdateGauge(ctx context.Context, name string, value float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.gauges[name] = value
	return nil
}

func (s *MemStorage) UpdateCounter(ctx context.Context, name string, delta int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counters[name] += delta
	return nil
}
func (s *MemStorage) GetGauge(ctx context.Context, name string) (float64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if res, ok := s.gauges[name]; ok {
		return res, nil
	}
	return 0, errors.New("metric is not found")
}
func (s *MemStorage) GetCounter(ctx context.Context, name string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if res, ok := s.counters[name]; ok {
		return res, nil
	}
	return 0, errors.New("metric is not found")
}
func (s *MemStorage) GetAll(ctx context.Context) (map[string]float64, map[string]int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	g := make(map[string]float64, len(s.gauges))
	for k, v := range s.gauges {
		g[k] = v
	}

	c := make(map[string]int64, len(s.counters))
	for k, v := range s.counters {
		c[k] = v
	}
	return g, c, nil
}

func (s *MemStorage) SetAll(ctx context.Context, gauges map[string]float64, counters map[string]int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.gauges = make(map[string]float64, len(gauges))
	for k, v := range gauges {
		s.gauges[k] = v
	}

	s.counters = make(map[string]int64, len(counters))
	for k, v := range counters {
		s.counters[k] = v
	}

	return nil
}
func (s *MemStorage) UpdateBatch(ctx context.Context, metrics []models.Metrics) error {

	for _, v := range metrics {
		if v.ID == "" {
			return errors.New("id is empty")
		}
		switch v.MType {
		case models.Gauge:
			if v.Value == nil {
				return errors.New("value is nil")
			}
		case models.Counter:
			if v.Delta == nil {
				return errors.New("delta is nil")
			}
		default:
			return errors.New("invalid metric type")
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, v := range metrics {
		switch v.MType {
		case models.Gauge:
			s.gauges[v.ID] = *v.Value
		case models.Counter:
			s.counters[v.ID] += *v.Delta
		}
	}

	return nil
}
func (s *MemStorage) Ping(ctx context.Context) error {
	return nil
}
