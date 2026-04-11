package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/config"
)

func parseFlags() config.ClientConfig {
	var (
		reportIntervalSec int
		pollIntervalSec   int
		cfg               config.ClientConfig
	)

	flag.StringVar(&cfg.Addr, "a", "localhost:8080", "server address")
	flag.IntVar(&reportIntervalSec, "r", 10, "report interval in seconds")
	flag.IntVar(&pollIntervalSec, "p", 2, "poll interval in seconds")
	flag.StringVar(&cfg.Key, "k", "", "hash key")
	flag.Parse()

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		cfg.Addr = envAddr
	}

	if envReport := os.Getenv("REPORT_INTERVAL"); envReport != "" {
		sec, err := strconv.Atoi(envReport)
		if err != nil {
			log.Fatal(err)
		}
		reportIntervalSec = sec
	}

	if envPoll := os.Getenv("POLL_INTERVAL"); envPoll != "" {
		sec, err := strconv.Atoi(envPoll)
		if err != nil {
			log.Fatal(err)
		}
		pollIntervalSec = sec
	}
	if envKey := os.Getenv("KEY"); envKey != "" {
		cfg.Key = envKey
	}

	cfg.ReportInterval = time.Duration(reportIntervalSec) * time.Second
	cfg.PollInterval = time.Duration(pollIntervalSec) * time.Second

	return cfg
}
