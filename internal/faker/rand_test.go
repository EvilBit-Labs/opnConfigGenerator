package faker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRandDeterministic(t *testing.T) {
	t.Parallel()

	a, fa := newRand(42)
	b, fb := newRand(42)
	require.NotNil(t, a)
	require.NotNil(t, b)
	require.NotNil(t, fa)
	require.NotNil(t, fb)

	for range 5 {
		assert.Equal(t, a.Uint64(), b.Uint64(), "same seed must produce identical PCG streams")
	}
}

func TestNewRandZeroSeedIsRandom(t *testing.T) {
	t.Parallel()

	a, _ := newRand(0)
	b, _ := newRand(0)
	assert.NotEqual(t, a.Uint64(), b.Uint64(), "zero-seed streams must diverge")
}

func TestNewRandGofakeitHonorsSeed(t *testing.T) {
	t.Parallel()

	_, fa := newRand(17)
	_, fb := newRand(17)
	assert.Equal(t, fa.Name(), fb.Name(), "gofakeit sharing the rand stream must be deterministic")
}
