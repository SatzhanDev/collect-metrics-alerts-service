package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type ServerFileConfig struct {
	Address       *string `json:"address"`
	Restore       *bool   `json:"restore"`
	StoreInterval *string `json:"store_interval"`
	StoreFile     *string `json:"store_file"`
	DatabaseDSN   *string `json:"database_dsn"`
	CryptoKey     *string `json:"crypto_key"`
	LogLevel      *string `json:"log_level"`
	HashKey       *string `json:"hash_key"`
	AuditFile     *string `json:"audit_file"`
	AuditURL      *string `json:"audit_url"`
	TrustedSubnet *string `json:"trusted_subnet"`
}

type ClientFileConfig struct {
	Address        *string `json:"address"`
	ReportInterval *string `json:"report_interval"`
	PollInterval   *string `json:"poll_interval"`
	CryptoKey      *string `json:"crypto_key"`
	HashKey        *string `json:"hash_key"`
	RateLimit      *int    `json:"rate_limit"`
}

func loadFileConfig[T any](path string) (*T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg T
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	return &cfg, nil
}

func LoadServerFileConfig(path string) (*ServerFileConfig, error) {
	return loadFileConfig[ServerFileConfig](path)
}

func LoadClientFileConfig(path string) (*ClientFileConfig, error) {
	return loadFileConfig[ClientFileConfig](path)
}
