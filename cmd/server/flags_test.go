package main

import (
	"os"
	"testing"
	"time"

	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/config"
	"github.com/stretchr/testify/require"
)

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }

func TestApplyServerFileConfig_FillsUnsetOptions(t *testing.T) {
	cfg := config.ServerConfig{
		Addr:     "localhost:8080",
		LogLevel: "info",
	}
	file := &config.ServerFileConfig{
		Address:       strPtr("localhost:9090"),
		Restore:       boolPtr(true),
		StoreInterval: strPtr("1s"),
		StoreFile:     strPtr("/tmp/metrics.db"),
		DatabaseDSN:   strPtr("postgres://localhost/db"),
		CryptoKey:     strPtr("/path/priv.pem"),
		LogLevel:      strPtr("debug"),
		HashKey:       strPtr("filesecret"),
		AuditFile:     strPtr("/tmp/audit.log"),
		AuditURL:      strPtr("http://localhost/audit"),
	}

	applyServerFileConfig(&cfg, file, map[string]bool{})

	require.Equal(t, "localhost:9090", cfg.Addr)
	require.True(t, cfg.Restore)
	require.Equal(t, 1*time.Second, cfg.StoreInterval)
	require.Equal(t, "/tmp/metrics.db", cfg.FileStoragePath)
	require.Equal(t, "postgres://localhost/db", cfg.DBDSN)
	require.Equal(t, "/path/priv.pem", cfg.CryptoKey)
	require.Equal(t, "debug", cfg.LogLevel)
	require.Equal(t, "filesecret", cfg.Key)
	require.Equal(t, "/tmp/audit.log", cfg.AuditFile)
	require.Equal(t, "http://localhost/audit", cfg.AuditURL)
}

func TestApplyServerFileConfig_ExplicitFlagWins(t *testing.T) {
	cfg := config.ServerConfig{Addr: "localhost:1111"}
	file := &config.ServerFileConfig{Address: strPtr("localhost:9090")}

	applyServerFileConfig(&cfg, file, map[string]bool{"a": true})

	require.Equal(t, "localhost:1111", cfg.Addr)
}

func TestApplyServerFileConfig_EnvWins(t *testing.T) {
	cfg := config.ServerConfig{DBDSN: "postgres://from-env"}
	file := &config.ServerFileConfig{DatabaseDSN: strPtr("postgres://from-file")}

	require.NoError(t, os.Setenv("DATABASE_DSN", "postgres://from-env"))
	defer os.Unsetenv("DATABASE_DSN")

	applyServerFileConfig(&cfg, file, map[string]bool{})

	require.Equal(t, "postgres://from-env", cfg.DBDSN)
}

func TestApplyServerFileConfig_NilFieldsIgnored(t *testing.T) {
	cfg := config.ServerConfig{Addr: "localhost:8080", Restore: false}
	file := &config.ServerFileConfig{}

	applyServerFileConfig(&cfg, file, map[string]bool{})

	require.Equal(t, "localhost:8080", cfg.Addr)
	require.False(t, cfg.Restore)
}
