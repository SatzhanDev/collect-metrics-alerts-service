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
		configPath       string
	)

	flag.StringVar(&cfg.Addr, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.IntVar(&storeIntervalSec, "i", 300, "store interval in seconds")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "file storage path")
	flag.BoolVar(&cfg.Restore, "r", false, "restore metrics from file")
	flag.StringVar(&cfg.DBDSN, "d", "", "database dsn")
	flag.StringVar(&cfg.Key, "k", "", "hash key")
	flag.StringVar(&cfg.AuditFile, "audit-file", "", "audit file path")
	flag.StringVar(&cfg.AuditURL, "audit-url", "", "audit receiver url")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", "", "path to private key file for asymmetric encryption")
	flag.StringVar(&cfg.TrustedSubnet, "t", "", "trusted subnet in CIDR notation")
	flag.StringVar(&configPath, "c", "", "path to JSON config file")
	flag.StringVar(&configPath, "config", "", "path to JSON config file (alias for -c)")

	flag.Parse()
	cfg.StoreInterval = time.Duration(storeIntervalSec) * time.Second

	// explicitFlags хранит имена флагов, реально переданных в командной строке,
	// чтобы отличить их от значений по умолчанию — это нужно, чтобы значения
	// из файла конфигурации не перетирали ничего, что было задано явно.
	explicitFlags := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		explicitFlags[f.Name] = true
	})

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
	if envKey := os.Getenv("KEY"); envKey != "" {
		cfg.Key = envKey
	}

	if v := os.Getenv("AUDIT_FILE"); v != "" {
		cfg.AuditFile = v
	}
	if v := os.Getenv("AUDIT_URL"); v != "" {
		cfg.AuditURL = v
	}
	if v := os.Getenv("CRYPTO_KEY"); v != "" {
		cfg.CryptoKey = v
	}
	if v := os.Getenv("TRUSTED_SUBNET"); v != "" {
		cfg.TrustedSubnet = v
	}
	if v := os.Getenv("CONFIG"); v != "" {
		configPath = v
	}

	if configPath != "" {
		fileCfg, err := config.LoadServerFileConfig(configPath)
		if err != nil {
			log.Fatal(err)
		}
		applyServerFileConfig(&cfg, fileCfg, explicitFlags)
	}

	return cfg
}

// applyServerFileConfig подставляет значения из файла конфигурации только
// туда, где опция не была явно задана ни флагом, ни переменной окружения
func applyServerFileConfig(cfg *config.ServerConfig, file *config.ServerFileConfig, explicitFlags map[string]bool) {
	config.MergeField(explicitFlags, "ADDRESS", "a", file.Address, func(v string) { cfg.Addr = v })
	config.MergeField(explicitFlags, "LOG_LEVEL", "l", file.LogLevel, func(v string) { cfg.LogLevel = v })

	config.MergeField(explicitFlags, "STORE_INTERVAL", "i", file.StoreInterval, func(v string) {
		d, err := time.ParseDuration(v)
		if err != nil {
			log.Fatal(err)
		}
		cfg.StoreInterval = d
	})

	config.MergeField(explicitFlags, "FILE_STORAGE_PATH", "f", file.StoreFile, func(v string) { cfg.FileStoragePath = v })
	config.MergeField(explicitFlags, "RESTORE", "r", file.Restore, func(v bool) { cfg.Restore = v })
	config.MergeField(explicitFlags, "DATABASE_DSN", "d", file.DatabaseDSN, func(v string) { cfg.DBDSN = v })
	config.MergeField(explicitFlags, "KEY", "k", file.HashKey, func(v string) { cfg.Key = v })
	config.MergeField(explicitFlags, "AUDIT_FILE", "audit-file", file.AuditFile, func(v string) { cfg.AuditFile = v })
	config.MergeField(explicitFlags, "AUDIT_URL", "audit-url", file.AuditURL, func(v string) { cfg.AuditURL = v })
	config.MergeField(explicitFlags, "CRYPTO_KEY", "crypto-key", file.CryptoKey, func(v string) { cfg.CryptoKey = v })
	config.MergeField(explicitFlags, "TRUSTED_SUBNET", "t", file.TrustedSubnet, func(v string) { cfg.TrustedSubnet = v })
}
