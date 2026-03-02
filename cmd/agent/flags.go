package main

import (
	"flag"
)

type Config struct {
	Addr           string
	ReportInterval int
	PollInterval   int
}

func parseFlags() Config {
	var cfg Config

	flag.StringVar(&cfg.Addr, "a", "localhost:8080", "server address")
	flag.IntVar(&cfg.ReportInterval, "r", 10, "report interval")
	flag.IntVar(&cfg.PollInterval, "p", 2, "poll metrics interval")

	flag.Parse()
	return cfg
}
