package main

import (
	"time"

	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/agent"
)

func main() {
	storage := agent.NewMetricsStorage()
	sender := agent.NewHTTPSender("http://localhost:8080")
	a := agent.NewAgent(storage, sender)

	go func() {
		for {
			a.Poll()
			time.Sleep(2 * time.Second)
		}
	}()

	for {
		a.Report()
		time.Sleep(10 * time.Second)
	}
}
