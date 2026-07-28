package agent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	models "github.com/SatzhanDev/collect-metrics-alerts-service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockSender struct {
	gauges   map[string]float64
	counters map[string]int64
}

func NewMockSender() *MockSender {
	return &MockSender{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (m *MockSender) SendGauge(ctx context.Context, name string, value float64) error {
	m.gauges[name] = value
	return nil
}

func (m *MockSender) SendCounter(ctx context.Context, name string, value int64) error {
	m.counters[name] = value
	return nil
}
func (m *MockSender) SendGaugeJSON(ctx context.Context, name string, value float64) error {
	m.gauges[name] = value
	return nil
}

func (m *MockSender) SendCounterJSON(ctx context.Context, name string, value int64) error {
	m.counters[name] = value
	return nil
}
func (m *MockSender) SendBatch(ctx context.Context, metrics []models.Metrics) error {
	for _, metric := range metrics {
		switch metric.MType {
		case models.Gauge:
			if metric.Value != nil {
				m.gauges[metric.ID] = *metric.Value
			}
		case models.Counter:
			if metric.Delta != nil {
				m.counters[metric.ID] = *metric.Delta
			}
		}
	}
	return nil
}
func TestAgent_Poll_IncrementsPollCount_AndSetsRandomValue(t *testing.T) {
	st := NewMetricsStorage()
	mock := NewMockSender()
	a := NewAgent(st, mock, 1)

	a.Poll()

	gauges, counters := st.Snapshot()

	assert.Equal(t, int64(1), counters["PollCount"])
	assert.Contains(t, gauges, "RandomValue")

}
func TestAgent_CollectRuntimeBatch_AfterPoll(t *testing.T) {
	st := NewMetricsStorage()
	mock := NewMockSender()
	a := NewAgent(st, mock, 1)

	a.Poll()

	batch := a.CollectRuntimeBatch()

	assert.NotEmpty(t, batch)

	var foundRandom, foundPollCount bool
	for _, m := range batch {
		if m.ID == "RandomValue" && m.MType == models.Gauge {
			foundRandom = true
		}
		if m.ID == "PollCount" && m.MType == models.Counter {
			foundPollCount = true
		}
	}
	assert.True(t, foundRandom, "должен содержать RandomValue")
	assert.True(t, foundPollCount, "должен содержать PollCount")
}

func TestAgent_CollectRuntimeBatch_Empty(t *testing.T) {
	st := NewMetricsStorage()
	mock := NewMockSender()
	a := NewAgent(st, mock, 1)

	// Без Poll хранилище пустое
	batch := a.CollectRuntimeBatch()
	assert.Empty(t, batch)
}

func TestAgent_Stop_NoWorkers(t *testing.T) {
	st := NewMetricsStorage()
	mock := NewMockSender()
	a := NewAgent(st, mock, 1)

	// Stop без запущенных воркеров не должен паниковать
	a.Stop()
}

func TestAgent_Stop_DeliversQueuedBatchAfterShutdownSignal(t *testing.T) {
	var received int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&received, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewHTTPSender(server.URL, "")
	st := NewMetricsStorage()
	a := NewAgent(st, sender, 1)

	// Как в исправленном cmd/agent/main.go: воркеры отправляют через отдельный,
	// не отменяемый контекст, а не через тот, что отменяется по сигналу
	// завершения — иначе уже поставленный в очередь батч не смог бы уйти.
	a.StartWorkers(context.Background(), 1)

	value := 1.0
	batch := []models.Metrics{{ID: "cpu", MType: models.Gauge, Value: &value}}
	a.jobs <- batch

	// Имитируем получение сигнала завершения: Stop() ждёт продюсеров (их нет),
	// закрывает канал jobs и ждёт, пока воркер дошлёт уже поставленный батч.
	a.Stop()

	assert.Equal(t, int32(1), atomic.LoadInt32(&received))
}

func TestSendBatch_FailsWithAlreadyCanceledContext(t *testing.T) {
	// Показывает, почему воркерам нельзя передавать отменяемый контекст:
	// если контекст уже отменён (как ctx после cancel() в main.go при
	// получении сигнала), запрос не уйдёт вообще — сразу вернётся ошибка.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewHTTPSender(server.URL, "")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // отменяем контекст ДО отправки — как если бы shutdown уже начался

	value := 1.0
	metrics := []models.Metrics{{ID: "cpu", MType: models.Gauge, Value: &value}}

	// sendOnce вместо SendBatch, чтобы не ждать retry-задержки: ошибка
	// "context canceled" не исчезнет ни на одной попытке.
	err := sender.sendOnce(ctx, metrics)
	require.Error(t, err)
}

func TestAgent_CollectSystemBatch(t *testing.T) {
	st := NewMetricsStorage()
	mock := NewMockSender()
	a := NewAgent(st, mock, 1)

	batch, err := a.CollectSystemBatch(t.Context())
	require.NoError(t, err)
	assert.NotEmpty(t, batch)

	// Должны быть TotalMemory и FreeMemory
	var hasTotal, hasFree bool
	for _, m := range batch {
		if m.ID == "TotalMemory" {
			hasTotal = true
		}
		if m.ID == "FreeMemory" {
			hasFree = true
		}
	}
	assert.True(t, hasTotal, "должен содержать TotalMemory")
	assert.True(t, hasFree, "должен содержать FreeMemory")
}

func TestAgent_Report_SendsSnapshotMetrics(t *testing.T) {
	st := NewMetricsStorage()
	mock := NewMockSender()
	a := NewAgent(st, mock, 1)

	_ = st.SetGauge("Alloc", 123.45)
	_ = st.SetGauge("RandomValue", 0.99)
	_ = st.AddCounter("PollCount", 7)

	a.Report(t.Context())
	require.Contains(t, mock.gauges, "Alloc")
	assert.Equal(t, 123.45, mock.gauges["Alloc"])

	require.Contains(t, mock.gauges, "RandomValue")
	assert.Equal(t, 0.99, mock.gauges["RandomValue"])

	require.Contains(t, mock.counters, "PollCount")
	assert.Equal(t, int64(7), mock.counters["PollCount"])

}
