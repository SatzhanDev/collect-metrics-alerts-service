// Package config содержит конфигурационные структуры сервера и агента.
package config

import "time"

type ServerConfig struct {
	Addr            string
	GRPCAddr        string
	LogLevel        string
	FileStoragePath string
	StoreInterval   time.Duration
	Restore         bool
	DBDSN           string
	Key             string
	AuditFile       string
	AuditURL        string
	CryptoKey       string
	TrustedSubnet   string
}

type ClientConfig struct {
	Addr           string
	GRPCAddr       string
	ReportInterval time.Duration
	PollInterval   time.Duration
	RateLimit      int
	Key            string
	CryptoKey      string
}
