package audit

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileObserver_Handle(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "audit.log")

	observer := NewFileObserver(tmpFile)

	event := Event{
		TS:        123,
		Metrics:   []string{"Alloc"},
		IPAddress: "127.0.0.1",
	}

	err := observer.Handle(context.Background(), event)

	require.NoError(t, err)

	data, err := os.ReadFile(tmpFile)

	require.NoError(t, err)
	require.Contains(t, string(data), "Alloc")
}
