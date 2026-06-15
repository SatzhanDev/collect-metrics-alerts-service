// Package models содержит типы данных, используемые в сервисе сбора метрик.
package models

const (
	// Counter — тип метрики для накапливаемых целочисленных значений.
	Counter = "counter"
	// Gauge — тип метрики для текущих значений с плавающей точкой.
	Gauge = "gauge"
)

// Metrics представляет единицу метрики, передаваемую между агентом и сервером.
// Delta и Value объявлены через указатели, чтобы отличать значение "0" от незаданного.
type Metrics struct {
	ID    string   `json:"id"`              // ID — имя метрики.
	MType string   `json:"type"`            // MType — тип метрики: counter или gauge.
	Delta *int64   `json:"delta,omitempty"` // Delta — значение counter-метрики.
	Value *float64 `json:"value,omitempty"` // Value — значение gauge-метрики.
	Hash  string   `json:"hash,omitempty"`  // Hash — контрольная сумма для проверки целостности.
}
