package agent

import (
	"context"
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

func (m *MockSender) SendGauge(name string, value float64) error {
	m.gauges[name] = value
	return nil
}

func (m *MockSender) SendCounter(name string, value int64) error {
	m.counters[name] = value
	return nil
}
func (m *MockSender) SendGaugeJSON(name string, value float64) error {
	m.gauges[name] = value
	return nil
}

func (m *MockSender) SendCounterJSON(name string, value int64) error {
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
