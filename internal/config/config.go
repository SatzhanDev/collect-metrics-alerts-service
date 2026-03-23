package config

import "time"

type ServerConfig struct {
	Addr          string
	LogLevel      string
	FilePath      string
	StoreInterval time.Duration
	Restore       bool
}

type ClientConfig struct {
	Addr           string
	ReportInterval time.Duration
	PollInterval   time.Duration
}
