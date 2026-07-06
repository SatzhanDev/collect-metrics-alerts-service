package main

import (
	"os"
	"testing"
	"time"

	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/config"
	"github.com/stretchr/testify/require"
)

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }

func TestApplyClientFileConfig_FillsUnsetOptions(t *testing.T) {
	cfg := config.ClientConfig{
		Addr:      "localhost:8080",
		Key:       "",
		RateLimit: 1,
	}
	file := &config.ClientFileConfig{
		Address:        strPtr("localhost:9090"),
		ReportInterval: strPtr("5s"),
		PollInterval:   strPtr("1s"),
		HashKey:        strPtr("filesecret"),
		RateLimit:      intPtr(3),
		CryptoKey:      strPtr("/path/pub.pem"),
	}

	applyClientFileConfig(&cfg, file, map[string]bool{})

	require.Equal(t, "localhost:9090", cfg.Addr)
	require.Equal(t, 5*time.Second, cfg.ReportInterval)
	require.Equal(t, 1*time.Second, cfg.PollInterval)
	require.Equal(t, "filesecret", cfg.Key)
	require.Equal(t, 3, cfg.RateLimit)
	require.Equal(t, "/path/pub.pem", cfg.CryptoKey)
}

func TestApplyClientFileConfig_ExplicitFlagWins(t *testing.T) {
	cfg := config.ClientConfig{Addr: "localhost:1111"}
	file := &config.ClientFileConfig{Address: strPtr("localhost:9090")}

	// "a" помечен как явно переданный флаг — значение из файла не должно применяться.
	applyClientFileConfig(&cfg, file, map[string]bool{"a": true})

	require.Equal(t, "localhost:1111", cfg.Addr)
}

func TestApplyClientFileConfig_EnvWins(t *testing.T) {
	cfg := config.ClientConfig{Addr: "localhost:1111"}
	file := &config.ClientFileConfig{Address: strPtr("localhost:9090")}

	require.NoError(t, os.Setenv("ADDRESS", "localhost:2222"))
	defer os.Unsetenv("ADDRESS")

	// cfg.Addr уже содержит значение, применённое из переменной окружения
	// на более раннем шаге parseFlags — applyClientFileConfig не должен его трогать.
	applyClientFileConfig(&cfg, file, map[string]bool{})

	require.Equal(t, "localhost:1111", cfg.Addr)
}

func TestApplyClientFileConfig_NilFieldsIgnored(t *testing.T) {
	cfg := config.ClientConfig{Addr: "localhost:8080", RateLimit: 1}
	file := &config.ClientFileConfig{}

	applyClientFileConfig(&cfg, file, map[string]bool{})

	require.Equal(t, "localhost:8080", cfg.Addr)
	require.Equal(t, 1, cfg.RateLimit)
}
