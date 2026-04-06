package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/config"
)

const (
	defaultMaxOpenConns    = 10
	defaultMaxIdleConns    = 5
	defaultConnMaxLifetime = 5 * time.Minute
)

func parseFlags() config.ServerConfig {
	var (
		cfg              config.ServerConfig
		storeIntervalSec int
	)

	flag.StringVar(&cfg.Addr, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.IntVar(&storeIntervalSec, "i", 300, "store interval in seconds")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "file storage path")
	flag.BoolVar(&cfg.Restore, "r", false, "restore metrics from file")
	flag.StringVar(&cfg.DBDSN, "d", "", "database dsn")

	flag.Parse()
	cfg.StoreInterval = time.Duration(storeIntervalSec) * time.Second

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		cfg.Addr = envAddr
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		cfg.LogLevel = envLogLevel
	}
	if envStoreIntStr := os.Getenv("STORE_INTERVAL"); envStoreIntStr != "" {
		sec, err := strconv.Atoi(envStoreIntStr)
		if err != nil {
			log.Fatal(err)
		}
		cfg.StoreInterval = time.Duration(sec) * time.Second
	}
	if envFilePath := os.Getenv("FILE_STORAGE_PATH"); envFilePath != "" {
		cfg.FileStoragePath = envFilePath
	}
	if envRestoreStr := os.Getenv("RESTORE"); envRestoreStr != "" {
		envRestore, err := strconv.ParseBool(envRestoreStr)
		if err != nil {
			log.Fatal(err)
		}
		cfg.Restore = envRestore
	}
	if envDBdsn := os.Getenv("DATABASE_DSN"); envDBdsn != "" {
		cfg.DBDSN = envDBdsn
	}

	return cfg
}
