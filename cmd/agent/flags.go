package main

import (
	"flag"
	"log"
	"os"
	"strconv"
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

	cfg.ReportInterval = time.Duration(reportIntervalSec) * time.Second
	cfg.PollInterval = time.Duration(pollIntervalSec) * time.Second

	return cfg
}
