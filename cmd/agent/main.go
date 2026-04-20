package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/agent"
)

func main() {
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
	fmt.Println("shutting down...")

	cancel()

	a.Wait()

	fmt.Println("graceful shutdown complete")

}
