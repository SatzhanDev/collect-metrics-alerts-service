package agent

import (
	"sync"
)

type MetricsStorage struct {
	gauges   map[string]float64
	counters map[string]int64
	mu       sync.RWMutex
}
type Storage interface {
	SetGauge(name string, v float64) error
	AddCounter(name string, delta int64) error
	Snapshot() (gauges map[string]float64, counters map[string]int64)
}

func NewMetricsStorage() *MetricsStorage {
	return &MetricsStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (s *MetricsStorage) SetGauge(name string, v float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.gauges[name] = v
	return nil
}

func (s *MetricsStorage) AddCounter(name string, delta int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counters[name] += delta
	return nil
}

func (s *MetricsStorage) Snapshot() (map[string]float64, map[string]int64) {
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

	return g, c
}
