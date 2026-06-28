package agent

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	models "github.com/SatzhanDev/collect-metrics-alerts-service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPSender_SendGauge_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "gauge")
		assert.Contains(t, r.URL.Path, "cpu")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewHTTPSender(server.URL, "")
	err := sender.SendGauge("cpu", 12.5)
	require.NoError(t, err)
}

func TestHTTPSender_SendGauge_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	sender := NewHTTPSender(server.URL, "")
	err := sender.SendGauge("cpu", 12.5)
	require.Error(t, err)
}

func TestHTTPSender_SendGaugeJSON_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var m models.Metrics
		err := json.NewDecoder(r.Body).Decode(&m)
		require.NoError(t, err)
		assert.Equal(t, "cpu", m.ID)
		assert.Equal(t, models.Gauge, m.MType)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewHTTPSender(server.URL, "")
	err := sender.SendGaugeJSON("cpu", 12.5)
	require.NoError(t, err)
}

func TestHTTPSender_SendGaugeJSON_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	sender := NewHTTPSender(server.URL, "")
	err := sender.SendGaugeJSON("cpu", 99.9)
	require.Error(t, err)
}

func TestHTTPSender_SendCounter_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "counter")
		assert.Contains(t, r.URL.Path, "PollCount")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewHTTPSender(server.URL, "")
	err := sender.SendCounter("PollCount", 5)
	require.NoError(t, err)
}

func TestHTTPSender_SendCounter_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	sender := NewHTTPSender(server.URL, "")
	err := sender.SendCounter("PollCount", 5)
	require.Error(t, err)
}

func TestHTTPSender_SendCounterJSON_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var m models.Metrics
		err := json.NewDecoder(r.Body).Decode(&m)
		require.NoError(t, err)
		assert.Equal(t, models.Counter, m.MType)
		assert.Equal(t, "PollCount", m.ID)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewHTTPSender(server.URL, "")
	err := sender.SendCounterJSON("PollCount", 5)
	require.NoError(t, err)
}

func TestHTTPSender_SendCounterJSON_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	sender := NewHTTPSender(server.URL, "")
	err := sender.SendCounterJSON("PollCount", 1)
	require.Error(t, err)
}

func TestHTTPSender_SendBatch_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		gr, err := gzip.NewReader(r.Body)
		require.NoError(t, err)
		defer gr.Close()

		body, err := io.ReadAll(gr)
		require.NoError(t, err)

		var metrics []models.Metrics
		require.NoError(t, json.Unmarshal(body, &metrics))
		assert.Len(t, metrics, 2)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewHTTPSender(server.URL, "")

	value := 12.5
	delta := int64(3)
	metrics := []models.Metrics{
		{ID: "cpu", MType: models.Gauge, Value: &value},
		{ID: "PollCount", MType: models.Counter, Delta: &delta},
	}

	err := sender.SendBatch(context.Background(), metrics)
	require.NoError(t, err)
}

func TestHTTPSender_SendBatch_Empty(t *testing.T) {
	// Не должны делать HTTP-запрос при пустом батче
	sender := NewHTTPSender("http://127.0.0.1:1", "")
	err := sender.SendBatch(context.Background(), nil)
	require.NoError(t, err)
}

func TestHTTPSender_SendBatch_WithKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hash := r.Header.Get("HashSHA256")
		assert.NotEmpty(t, hash)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewHTTPSender(server.URL, "secret")

	value := 42.0
	metrics := []models.Metrics{
		{ID: "test", MType: models.Gauge, Value: &value},
	}

	err := sender.SendBatch(context.Background(), metrics)
	require.NoError(t, err)
}

func TestHTTPSender_SendBatch_BadStatus(t *testing.T) {
	// Проверяем что ошибки сервера возвращаются
	// SendBatch делает 4 попытки, ждём быстрый ответ 400
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	sender := NewHTTPSender(server.URL, "")

	value := 1.0
	metrics := []models.Metrics{{ID: "cpu", MType: models.Gauge, Value: &value}}

	// Тестируем только sendOnce напрямую, чтобы не ждать retry-задержки
	err := sender.sendOnce(context.Background(), metrics)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bad status")
}

func TestHTTPSender_SendGauge_NoSchemeURL(t *testing.T) {
	// Пустой адрес → URL без схемы → ошибка "unsupported protocol scheme"
	sender := NewHTTPSender("", "")
	err := sender.SendGauge("cpu", 1.0)
	require.Error(t, err)
}

func TestHTTPSender_SendBatch_WithoutHashKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.Header.Get("HashSHA256"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewHTTPSender(server.URL, "")

	value := 1.0
	metrics := []models.Metrics{{ID: "x", MType: models.Gauge, Value: &value}}

	err := sender.sendOnce(context.Background(), metrics)
	require.NoError(t, err)
}

func TestHTTPSender_SendGaugeJSON_URL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.True(t, strings.HasSuffix(r.URL.Path, "/update"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewHTTPSender(server.URL, "")
	err := sender.SendGaugeJSON("mem", 256.0)
	require.NoError(t, err)
}
