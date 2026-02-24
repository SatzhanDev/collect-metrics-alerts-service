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

	// if got := counters["PollCount"]; got != 1 {
	// 	t.Fatalf("expected PollCount=1, got %d", got)
	// }
	// if _, ok := gauges["RandomValue"]; !ok {
	// 	t.Fatalf("expected RandomValue to be set")
	// }
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

	// if got, ok := mock.gauges["Alloc"]; !ok || got != 123.45 {
	// 	t.Fatalf("expected gauge Alloc=123.45, got %v (exists=%v)", got, ok)
	// }
	// if got, ok := mock.gauges["RandomValue"]; !ok || got != 0.99 {
	// 	t.Fatalf("expected gauge RandomValue=0.99, got %v (exists=%v)", got, ok)
	// }
	// if got, ok := mock.counters["PollCount"]; !ok || got != 7 {
	// 	t.Fatalf("expected counter PollCount=7, got %v (exists=%v)", got, ok)
	// }
}
