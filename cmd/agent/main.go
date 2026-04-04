package main

import (
	"context"
	"time"

	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/agent"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/logger"
	"go.uber.org/zap"
)

func main() {
	cfg := parseFlags()

	storage := agent.NewMetricsStorage()
	sender := agent.NewHTTPSender("http://" + cfg.Addr)
	a := agent.NewAgent(storage, sender)

	go func() {
		for {
			a.Poll()
			time.Sleep(cfg.PollInterval)
		}
	}()

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		if err := a.Report(ctx); err != nil {
			logger.Log.Error("failed to report metric", zap.Error(err))

		}
		cancel()
		time.Sleep(cfg.ReportInterval)
	}
}
