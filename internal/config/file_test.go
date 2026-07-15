package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadServerFileConfig(t *testing.T) {
	content := `{
		"address": "localhost:9090",
		"restore": true,
		"store_interval": "1s",
		"store_file": "/tmp/metrics.db",
		"database_dsn": "postgres://localhost/db",
		"crypto_key": "/tmp/priv.pem",
		"log_level": "debug",
		"hash_key": "secret",
		"audit_file": "/tmp/audit.log",
		"audit_url": "http://localhost/audit",
		"trusted_subnet": "192.168.1.0/24"
	}`

	path := filepath.Join(t.TempDir(), "server.json")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	cfg, err := LoadServerFileConfig(path)
	require.NoError(t, err)

	require.Equal(t, "localhost:9090", *cfg.Address)
	require.True(t, *cfg.Restore)
	require.Equal(t, "1s", *cfg.StoreInterval)
	require.Equal(t, "/tmp/metrics.db", *cfg.StoreFile)
	require.Equal(t, "postgres://localhost/db", *cfg.DatabaseDSN)
	require.Equal(t, "/tmp/priv.pem", *cfg.CryptoKey)
	require.Equal(t, "debug", *cfg.LogLevel)
	require.Equal(t, "secret", *cfg.HashKey)
	require.Equal(t, "/tmp/audit.log", *cfg.AuditFile)
	require.Equal(t, "http://localhost/audit", *cfg.AuditURL)
	require.Equal(t, "192.168.1.0/24", *cfg.TrustedSubnet)
}

func TestLoadServerFileConfig_MissingFieldsStayNil(t *testing.T) {
	path := filepath.Join(t.TempDir(), "server.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"address": "localhost:9090"}`), 0o600))

	cfg, err := LoadServerFileConfig(path)
	require.NoError(t, err)

	require.NotNil(t, cfg.Address)
	require.Nil(t, cfg.Restore)
	require.Nil(t, cfg.StoreInterval)
	require.Nil(t, cfg.StoreFile)
	require.Nil(t, cfg.DatabaseDSN)
	require.Nil(t, cfg.CryptoKey)
	require.Nil(t, cfg.TrustedSubnet)
}

func TestLoadServerFileConfig_FileNotFound(t *testing.T) {
	_, err := LoadServerFileConfig(filepath.Join(t.TempDir(), "missing.json"))
	require.Error(t, err)
}

func TestLoadServerFileConfig_InvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "server.json")
	require.NoError(t, os.WriteFile(path, []byte("not json"), 0o600))

	_, err := LoadServerFileConfig(path)
	require.Error(t, err)
}

func TestLoadClientFileConfig(t *testing.T) {
	content := `{
		"address": "localhost:9090",
		"report_interval": "2s",
		"poll_interval": "500ms",
		"crypto_key": "/tmp/pub.pem",
		"hash_key": "secret",
		"rate_limit": 5
	}`

	path := filepath.Join(t.TempDir(), "agent.json")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	cfg, err := LoadClientFileConfig(path)
	require.NoError(t, err)

	require.Equal(t, "localhost:9090", *cfg.Address)
	require.Equal(t, "2s", *cfg.ReportInterval)
	require.Equal(t, "500ms", *cfg.PollInterval)
	require.Equal(t, "/tmp/pub.pem", *cfg.CryptoKey)
	require.Equal(t, "secret", *cfg.HashKey)
	require.Equal(t, 5, *cfg.RateLimit)
}

func TestLoadClientFileConfig_MissingFieldsStayNil(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agent.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"address": "localhost:9090"}`), 0o600))

	cfg, err := LoadClientFileConfig(path)
	require.NoError(t, err)

	require.NotNil(t, cfg.Address)
	require.Nil(t, cfg.ReportInterval)
	require.Nil(t, cfg.PollInterval)
	require.Nil(t, cfg.RateLimit)
}

func TestLoadClientFileConfig_FileNotFound(t *testing.T) {
	_, err := LoadClientFileConfig(filepath.Join(t.TempDir(), "missing.json"))
	require.Error(t, err)
}

func TestLoadClientFileConfig_InvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agent.json")
	require.NoError(t, os.WriteFile(path, []byte("not json"), 0o600))

	_, err := LoadClientFileConfig(path)
	require.Error(t, err)
}
