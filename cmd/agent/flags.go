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
	var cfg Config
	flag.StringVar(&cfg.Addr, "a", "localhost:8080", "server address")
	flag.Duration("r", cfg.ReportInterval, "report interval")
	flag.Duration("p", cfg.PollInterval, "poll metrics interval")
	flag.Parse()
	return cfg
}
