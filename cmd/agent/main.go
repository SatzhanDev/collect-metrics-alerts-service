package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/agent"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/logger"
)

// Переменные заполняются при сборке через -ldflags "-X main.buildVersion=v1.0.0 ..."
var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

// na возвращает значение переменной или "N/A" если она пустая.
func na(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}

func main() {
	fmt.Printf("Build version: %s\n", na(buildVersion))
	fmt.Printf("Build date: %s\n", na(buildDate))
	fmt.Printf("Build commit: %s\n", na(buildCommit))

	cfg := parseFlags()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	storage := agent.NewMetricsStorage()
	sender := agent.NewHTTPSender("http://"+cfg.Addr, cfg.Key)
	a := agent.NewAgent(storage, sender, cfg.RateLimit)

	a.StartWorkers(ctx, cfg.RateLimit)
	a.StartRuntimeCollector(ctx, cfg.PollInterval)
	a.StartSystemCollector(ctx, cfg.PollInterval)

	<-sigCh
	logger.Log.Info("shutting down agent")

	cancel()
	a.Stop()

	logger.Log.Info("graceful shutdown complete")

}
