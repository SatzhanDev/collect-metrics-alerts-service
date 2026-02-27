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
	s.gauges[name] = value
	return nil
}

func (s *MemStorage) UpdateCounter(name string, delta int64) error {
	s.counters[name] += delta
	return nil
}
func (s *MemStorage) GetGauge(name string) (float64, error) {
	if res, ok := s.gauges[name]; ok {
		return res, nil
	}
	return 0, errors.New("metric is not found")
}
func (s *MemStorage) GetCounter(name string) (int64, error) {
	if res, ok := s.counters[name]; ok {
		return res, nil
	}
	return 0, errors.New("metric is not found")
}
func (s *MemStorage) Snapshot() (gauges map[string]float64, counters map[string]int64) {
	return nil, nil
}
