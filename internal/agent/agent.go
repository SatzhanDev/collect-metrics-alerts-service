package agent

import (
	"context"
	"math/rand"

	models "github.com/SatzhanDev/collect-metrics-alerts-service/internal/model"
)

type Agent struct {
	storage Storage
	sender  Sender
}

func NewAgent(storage Storage, sender Sender) *Agent {
	return &Agent{
		storage: storage,
		sender:  sender,
	}
}

func (a *Agent) Poll() {
	gauges := CollectRuntimeMetrics()

	for name, value := range gauges {
		_ = a.storage.SetGauge(name, value)
	}

	_ = a.storage.AddCounter("PollCount", 1)

	randomValue := rand.Float64()
	_ = a.storage.SetGauge("RandomValue", randomValue)

}

func (a *Agent) Report(ctx context.Context) error {
	gauges, counters := a.storage.Snapshot()

	metrics := make([]models.Metrics, 0, len(gauges)+len(counters))
	for name, value := range gauges {
		v := value
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: &v,
		})
	}
	for name, delta := range counters {
		d := delta
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: models.Counter,
			Delta: &d,
		})
	}

	return a.sender.SendBatch(ctx, metrics)

}
