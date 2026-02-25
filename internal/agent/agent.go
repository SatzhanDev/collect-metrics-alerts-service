package agent

import "math/rand"

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

func (a *Agent) Report() {
	gauges, counters := a.storage.Snapshot()

	for name, value := range gauges {
		_ = a.sender.SendGauge(name, value)
	}

	for name, value := range counters {
		_ = a.sender.SendCounter(name, value)
	}
}
