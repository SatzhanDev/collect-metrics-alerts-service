package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/agent"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/buildinfo"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/cryptoutil"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/logger"
)

func main() {
	fmt.Printf("Build version: %s\n", buildinfo.NA(buildinfo.Version))
	fmt.Printf("Build date: %s\n", buildinfo.NA(buildinfo.Date))
	fmt.Printf("Build commit: %s\n", buildinfo.NA(buildinfo.Commit))

	cfg := parseFlags()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	storage := agent.NewMetricsStorage()
	sender := agent.NewHTTPSender("http://"+cfg.Addr, cfg.Key)

	if cfg.CryptoKey != "" {
		pubKey, err := cryptoutil.LoadPublicKey(cfg.CryptoKey)
		if err != nil {
			log.Fatal(err)
		}
		sender.SetPublicKey(pubKey)
	}

	a := agent.NewAgent(storage, sender, cfg.RateLimit)

	a.StartWorkers(context.Background(), cfg.RateLimit)
	a.StartRuntimeCollector(ctx, cfg.PollInterval)
	a.StartSystemCollector(ctx, cfg.PollInterval)

	<-sigCh
	logger.Log.Info("shutting down agent")

	cancel()
	a.Stop()

	logger.Log.Info("graceful shutdown complete")

}
