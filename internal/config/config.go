package config

import "time"

type ServerConfig struct {
	Addr            string
	LogLevel        string
	FileStoragePath string
	StoreInterval   time.Duration
	Restore         bool
	DBDSN           string
	Key             string
}

type ClientConfig struct {
	Addr           string
	ReportInterval time.Duration
	PollInterval   time.Duration
	RateLimit      int
	Key            string
}
