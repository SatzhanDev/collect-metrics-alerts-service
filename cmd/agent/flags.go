package main

import (
	"flag"
	"time"
)

type Config struct {
	Addr           string
	ReportInterval time.Duration
	PollInterval   time.Duration
}

func parseFlags() Config {
	var (
		reportIntervalSec int
		pollIntervalSec   int
		cfg               Config
	)

	flag.StringVar(&cfg.Addr, "a", "localhost:8080", "server address")
	flag.IntVar(&reportIntervalSec, "r", 10, "report interval in seconds")
	flag.IntVar(&pollIntervalSec, "p", 2, "poll interval in seconds")

	flag.Parse()

	cfg.ReportInterval = time.Duration(reportIntervalSec) * time.Second
	cfg.PollInterval = time.Duration(pollIntervalSec) * time.Second

	return cfg
}
