package repository

import (
	"errors"
	"sync"
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

func (s *MemStorage) UpdateGauge(name string, value float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.gauges[name] = value
	return nil
}

func (s *MemStorage) UpdateCounter(name string, delta int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counters[name] += delta
	return nil
}
func (s *MemStorage) GetGauge(name string) (float64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if res, ok := s.gauges[name]; ok {
		return res, nil
	}
	return 0, errors.New("metric is not found")
}
func (s *MemStorage) GetCounter(name string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if res, ok := s.counters[name]; ok {
		return res, nil
	}
	return 0, errors.New("metric is not found")
}
func (s *MemStorage) GetAll() (gauges map[string]float64, counters map[string]int64) {
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

func (s *MemStorage) SetAll(gauges map[string]float64, counters map[string]int64) error {
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
