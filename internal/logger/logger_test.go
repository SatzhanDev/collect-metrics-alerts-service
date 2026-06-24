package logger

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitialize_ValidLevel(t *testing.T) {
	err := Initialize("info")
	require.NoError(t, err)
	assert.NotNil(t, Log)
}

func TestInitialize_DebugLevel(t *testing.T) {
	err := Initialize("debug")
	require.NoError(t, err)
	assert.NotNil(t, Log)
}

func TestInitialize_InvalidLevel(t *testing.T) {
	err := Initialize("не-уровень-логирования")
	require.Error(t, err)
}

func TestWithLogging_Status200(t *testing.T) {
	require.NoError(t, Initialize("info"))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	wrapped := WithLogging(handler)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestWithLogging_Status500(t *testing.T) {
	require.NoError(t, Initialize("info"))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("error"))
	})

	wrapped := WithLogging(handler)

	req := httptest.NewRequest(http.MethodPost, "/update", nil)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestWithLogging_NoExplicitWriteHeader(t *testing.T) {
	require.NoError(t, Initialize("info"))

	// Если handler пишет только тело без явного WriteHeader,
	// loggingResponseWriter должен применить StatusOK по умолчанию
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("implicit 200"))
	})

	wrapped := WithLogging(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}
