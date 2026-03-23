package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/config"
)

func parseFlags() config.ServerConfig {
	var (
		cfg              config.ServerConfig
		storeIntervalSec int
	)

	flag.StringVar(&cfg.Addr, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.StringVar(&cfg.FilePath, "f", "metricsFile.txt", "file storage path")
	flag.IntVar(&storeIntervalSec, "i", 300, "store interval in seconds")
	flag.BoolVar(&cfg.Restore, "r", false, "restore metrics from file")

	flag.Parse()

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		cfg.Addr = envAddr
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		cfg.LogLevel = envLogLevel
	}
	if envFileath := os.Getenv("FILE_STORAGE_PATH"); envFileath != "" {
		cfg.FilePath = envFileath
	}
	if envStoreIntStr := os.Getenv("STORE_INTERVAL"); envStoreIntStr != "" {
		sec, err := strconv.Atoi(envStoreIntStr)
		if err != nil {
			log.Fatal(err)
		}
		cfg.StoreInterval = time.Duration(sec) * time.Second
	}
	if envRestoreStr := os.Getenv("RESTORE"); envRestoreStr != "" {
		envRestore, err := strconv.ParseBool(envRestoreStr)
		if err != nil {
			log.Fatal(err)
		}
		cfg.Restore = envRestore
	}

	return cfg
}
