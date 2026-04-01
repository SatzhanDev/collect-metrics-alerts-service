package mem

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemStorage_UpdateGauge(t *testing.T) {
	s := NewMemStorage()

	err := s.UpdateGauge(t.Context(), "cpu", 10.5)
	require.NoError(t, err)

	val, err := s.GetGauge(t.Context(), "cpu")
	require.NoError(t, err)
	assert.Equal(t, 10.5, val)
}
