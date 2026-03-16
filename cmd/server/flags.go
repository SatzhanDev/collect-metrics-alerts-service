package main

import (
	"flag"
	"os"
)

type Config struct {
	Addr     string
	LogLevel string
}

func parseFlags() Config {
	var cfg Config
	flag.StringVar(&cfg.Addr, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")

	flag.Parse()

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		cfg.Addr = envAddr
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		cfg.LogLevel = envLogLevel
	}

	return cfg
}
