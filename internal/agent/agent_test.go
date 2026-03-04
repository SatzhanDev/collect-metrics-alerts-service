package agent

import (
	"testing"

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
func TestAgent_Poll_IncrementsPollCount_AndSetsRandomValue(t *testing.T) {
	st := NewMetricsStorage()
	mock := NewMockSender()
	a := NewAgent(st, mock)

	a.Poll()

	gauges, counters := st.Snapshot()

	assert.Equal(t, int64(1), counters["PollCount"])
	assert.Contains(t, gauges, "RandomValue")

}
func TestAgent_Report_SendsSnapshotMetrics(t *testing.T) {
	st := NewMetricsStorage()
	mock := NewMockSender()
	a := NewAgent(st, mock)

	_ = st.SetGauge("Alloc", 123.45)
	_ = st.SetGauge("RandomValue", 0.99)
	_ = st.AddCounter("PollCount", 7)

	a.Report()
	require.Contains(t, mock.gauges, "Alloc")
	assert.Equal(t, 123.45, mock.gauges["Alloc"])

	require.Contains(t, mock.gauges, "RandomValue")
	assert.Equal(t, 0.99, mock.gauges["RandomValue"])

	require.Contains(t, mock.counters, "PollCount")
	assert.Equal(t, int64(7), mock.counters["PollCount"])

}
