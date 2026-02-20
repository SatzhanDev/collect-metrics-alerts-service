package repository

type Storage interface {
	UpdateGauge(name string, value float64) error
	UpdateCounter(name string, delta int64) error
}
