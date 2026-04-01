package postgres

import (
	"context"
	"database/sql"
	"errors"
)

var ErrMetricNotFound = errors.New("metric is not found")

type Storage struct {
	db *sql.DB
}

func New(db *sql.DB) *Storage {
	return &Storage{
		db: db,
	}
}

func (s *Storage) UpdateGauge(ctx context.Context, name string, value float64) error {
	return nil
}

func (s *Storage) UpdateCounter(ctx context.Context, name string, delta int64) error {
	return nil
}
func (s *Storage) GetGauge(ctx context.Context, name string) (float64, error) {
	return 0, nil
}
func (s *Storage) GetCounter(ctx context.Context, name string) (int64, error) {
	return 0, nil
}
func (s *Storage) GetAll(ctx context.Context) (gauges map[string]float64, counters map[string]int64) {
	return nil, nil
}
func (s *Storage) SetAll(ctx context.Context, gauges map[string]float64, counters map[string]int64) error {
	return nil
}
