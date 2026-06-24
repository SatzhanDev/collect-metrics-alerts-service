package audit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPObserver_Handle_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	observer := NewHTTPObserver(server.URL)
	event := Event{
		TS:        1234567890,
		Metrics:   []string{"cpu", "mem"},
		IPAddress: "127.0.0.1",
	}

	err := observer.Handle(context.Background(), event)
	require.NoError(t, err)
}

func TestHTTPObserver_Handle_ConnectionRefused(t *testing.T) {
	// Используем порт 1, который точно недоступен
	observer := NewHTTPObserver("http://127.0.0.1:1")

	err := observer.Handle(context.Background(), Event{})
	require.Error(t, err)
}

func TestHTTPObserver_Handle_EmptyEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	observer := NewHTTPObserver(server.URL)
	err := observer.Handle(context.Background(), Event{})
	require.NoError(t, err)
}
