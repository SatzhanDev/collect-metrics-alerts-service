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
		configPath        string
	)

	flag.StringVar(&cfg.Addr, "a", "localhost:8080", "server address")
	flag.IntVar(&reportIntervalSec, "r", 10, "report interval in seconds")
	flag.IntVar(&pollIntervalSec, "p", 2, "poll interval in seconds")
	flag.StringVar(&cfg.Key, "k", "", "hash key")
	flag.IntVar(&cfg.RateLimit, "l", 1, "rate limit")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", "", "path to public key file for asymmetric encryption")
	flag.StringVar(&configPath, "c", "", "path to JSON config file")
	flag.StringVar(&configPath, "config", "", "path to JSON config file (alias for -c)")
	flag.Parse()

	explicitFlags := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		explicitFlags[f.Name] = true
	})

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
	if v := os.Getenv("RATE_LIMIT"); v != "" {
		val, _ := strconv.Atoi(v)
		cfg.RateLimit = val
	}
	if envCryptoKey := os.Getenv("CRYPTO_KEY"); envCryptoKey != "" {
		cfg.CryptoKey = envCryptoKey
	}
	if envConfig := os.Getenv("CONFIG"); envConfig != "" {
		configPath = envConfig
	}

	cfg.ReportInterval = time.Duration(reportIntervalSec) * time.Second
	cfg.PollInterval = time.Duration(pollIntervalSec) * time.Second

	if configPath != "" {
		fileCfg, err := config.LoadClientFileConfig(configPath)
		if err != nil {
			log.Fatal(err)
		}
		applyClientFileConfig(&cfg, fileCfg, explicitFlags)
	}

	return cfg
}

func applyClientFileConfig(cfg *config.ClientConfig, file *config.ClientFileConfig, explicitFlags map[string]bool) {
	config.MergeField(explicitFlags, "ADDRESS", "a", file.Address, func(v string) { cfg.Addr = v })

	config.MergeField(explicitFlags, "REPORT_INTERVAL", "r", file.ReportInterval, func(v string) {
		d, err := time.ParseDuration(v)
		if err != nil {
			log.Fatal(err)
		}
		cfg.ReportInterval = d
	})

	config.MergeField(explicitFlags, "POLL_INTERVAL", "p", file.PollInterval, func(v string) {
		d, err := time.ParseDuration(v)
		if err != nil {
			log.Fatal(err)
		}
		cfg.PollInterval = d
	})

	config.MergeField(explicitFlags, "KEY", "k", file.HashKey, func(v string) { cfg.Key = v })
	config.MergeField(explicitFlags, "RATE_LIMIT", "l", file.RateLimit, func(v int) { cfg.RateLimit = v })
	config.MergeField(explicitFlags, "CRYPTO_KEY", "crypto-key", file.CryptoKey, func(v string) { cfg.CryptoKey = v })
}
