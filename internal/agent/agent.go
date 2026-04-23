package agent

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/logger"
	models "github.com/SatzhanDev/collect-metrics-alerts-service/internal/model"
	"go.uber.org/zap"
)

type Agent struct {
	storage     Storage
	sender      Sender
	jobs        chan []models.Metrics
	producersWG sync.WaitGroup
	workersWG   sync.WaitGroup
}

func NewAgent(storage Storage, sender Sender, rateLimit int) *Agent {
	if rateLimit <= 0 {
		rateLimit = 1
	}
	return &Agent{
		storage: storage,
		sender:  sender,
		jobs:    make(chan []models.Metrics, rateLimit*2),
	}
}

func (a *Agent) runProducer(fn func()) {
	a.producersWG.Add(1)

	go func() {
		defer a.producersWG.Done()
		fn()
	}()
}

func (a *Agent) runWorker(fn func()) {
	a.workersWG.Add(1)

	go func() {
		defer a.workersWG.Done()
		fn()
	}()
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

func (a *Agent) CollectRuntimeBatch() []models.Metrics {
	gauges, counters := a.storage.Snapshot()

	var batch []models.Metrics

	for name, value := range gauges {
		v := value
		batch = append(batch, models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: &v,
		})
	}

	for name, delta := range counters {
		d := delta
		batch = append(batch, models.Metrics{
			ID:    name,
			MType: models.Counter,
			Delta: &d,
		})
	}
	return batch
}

func (a *Agent) StartWorkers(ctx context.Context, workers int) {
	for i := 0; i < workers; i++ {
		workerID := i

		a.runWorker(func() {
			for batch := range a.jobs {
				if len(batch) == 0 {
					continue
				}

				if err := a.sender.SendBatch(ctx, batch); err != nil {
					logger.Log.Error("send batch failed",
						zap.Int("worker_id", workerID),
						zap.Error(err),
					)
				}
			}

			logger.Log.Info("worker stopped", zap.Int("worker_id", workerID))
		})
	}
}

func (a *Agent) StartRuntimeCollector(ctx context.Context, interval time.Duration) {
	a.runProducer(func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				a.Poll()

				batch := a.CollectRuntimeBatch()
				if len(batch) > 0 {
					a.jobs <- batch
				}

			case <-ctx.Done():
				logger.Log.Info("runtime collector stopped")
				return
			}
		}
	})
}

func (a *Agent) StartSystemCollector(ctx context.Context, interval time.Duration) {
	a.runProducer(func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				batch, err := a.CollectSystemBatch(ctx)
				if err != nil {
					logger.Log.Error("system metrics failed", zap.Error(err))
					continue
				}

				if len(batch) > 0 {
					a.jobs <- batch
				}

			case <-ctx.Done():
				logger.Log.Info("system collector stopped")
				return
			}
		}
	})
}
func (a *Agent) Stop() {
	a.producersWG.Wait()

	close(a.jobs)

	a.workersWG.Wait()
}
