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
	// fmt.Println("Start:", cfg.Addr)

	go func() {
		for {
			a.Poll()
			time.Sleep(time.Duration(cfg.PollInterval) * time.Second)
		}
	}()

	for {
		a.Report()
		time.Sleep(time.Duration(cfg.ReportInterval) * time.Second)
	}
}
