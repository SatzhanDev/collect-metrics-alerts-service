package audit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type mockObserver struct {
	called bool
}

func (m *mockObserver) Handle(ctx context.Context, event Event) error {
	m.called = true
	return nil
}

func TestPublisher_Notify(t *testing.T) {
	pub := NewPublisher(zap.NewNop())

	mock := &mockObserver{}

	pub.Subscribe(mock)

	pub.Notify(context.Background(), Event{})

	require.True(t, mock.called)
}

func TestPublisher_Enabled_NoObservers(t *testing.T) {
	pub := NewPublisher(zap.NewNop())
	require.False(t, pub.Enabled())
}

func TestPublisher_Enabled_WithObserver(t *testing.T) {
	pub := NewPublisher(zap.NewNop())
	pub.Subscribe(&mockObserver{})
	require.True(t, pub.Enabled())
}
