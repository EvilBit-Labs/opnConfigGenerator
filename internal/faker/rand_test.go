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

	// Sampling a single Uint64 pair has a ~1 in 2^64 collision chance —
	// astronomical but not zero. Checking 8 consecutive draws and asserting
	// at least one differs brings flake probability to 2^-512.
	const samples = 8
	diverged := false
	for range samples {
		if a.Uint64() != b.Uint64() {
			diverged = true
			break
		}
	}
	assert.True(t, diverged, "zero-seed streams must diverge within %d draws", samples)
}

func TestNewRandGofakeitHonorsSeed(t *testing.T) {
	t.Parallel()

	_, fa := newRand(17)
	_, fb := newRand(17)
	assert.Equal(t, fa.Name(), fb.Name(), "gofakeit sharing the rand stream must be deterministic")
}
