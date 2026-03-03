package main

import (
	"time"

	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/agent"
)

func main() {
	cfg := parseFlags()

	storage := agent.NewMetricsStorage()
	sender := agent.NewHTTPSender("http://" + cfg.Addr)
	a := agent.NewAgent(storage, sender)

	go func() {
		for {
			a.Poll()
			time.Sleep(cfg.PollInterval * time.Second)
		}
	}()

	for {
		a.Report()
		time.Sleep(cfg.ReportInterval * time.Second)
	}
}
