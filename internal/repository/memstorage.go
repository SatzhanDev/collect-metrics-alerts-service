package repository

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

func NewMemStorage() MemStorage {
	return MemStorage{
		gauges:   map[string]float64{},
		counters: map[string]int64{},
	}
}

func (s *MemStorage) SetGauge(name string, value float64) error {
	return nil
}

func (s *MemStorage) AddCounter(name string, delta int64) error {
	return nil
}
