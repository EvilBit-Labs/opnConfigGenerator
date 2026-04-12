package generator_test

import (
	"testing"

	"github.com/EvilBit-Labs/opnConfigGenerator/internal/generator"
	"github.com/EvilBit-Labs/opnConfigGenerator/internal/netutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSingle(t *testing.T) {
	t.Parallel()

	gen := generator.NewVlanGenerator(int64Ptr(42), generator.WanSingle)
	cfg, err := gen.GenerateSingle()
	require.NoError(t, err)

	assert.GreaterOrEqual(t, cfg.VlanID, uint16(generator.MinVlanID))
	assert.LessOrEqual(t, cfg.VlanID, uint16(generator.MaxVlanID))
	assert.True(t, netutil.IsRFC1918(cfg.IPNetwork), "network should be RFC 1918")
	assert.Equal(t, 24, cfg.IPNetwork.Bits(), "network should be /24")
	assert.NotEmpty(t, cfg.Description)
	assert.Equal(t, uint8(1), cfg.WanAssignment)
}

func TestGenerateBatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		count   int
		wantErr bool
	}{
		{"single", 1, false},
		{"ten", 10, false},
		{"hundred", 100, false},
		{"zero count", 0, true},
		{"negative count", -1, true},
		{"exceeds max", generator.MaxUniqueVlans + 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gen := generator.NewVlanGenerator(int64Ptr(42), generator.WanSingle)
			configs, err := gen.GenerateBatch(tt.count)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, configs, tt.count)
		})
	}
}

func TestVlanIDUniqueness(t *testing.T) {
	t.Parallel()

	gen := generator.NewVlanGenerator(int64Ptr(42), generator.WanSingle)
	configs, err := gen.GenerateBatch(500)
	require.NoError(t, err)

	seen := make(map[uint16]bool)
	for _, cfg := range configs {
		assert.False(t, seen[cfg.VlanID], "duplicate VLAN ID: %d", cfg.VlanID)
		seen[cfg.VlanID] = true
	}
}

func TestNetworkUniqueness(t *testing.T) {
	t.Parallel()

	gen := generator.NewVlanGenerator(int64Ptr(42), generator.WanSingle)
	configs, err := gen.GenerateBatch(500)
	require.NoError(t, err)

	seen := make(map[string]bool)
	for _, cfg := range configs {
		key := cfg.IPNetwork.String()
		assert.False(t, seen[key], "duplicate network: %s", key)
		seen[key] = true
	}
}

func TestAllNetworksRFC1918(t *testing.T) {
	t.Parallel()

	gen := generator.NewVlanGenerator(int64Ptr(42), generator.WanSingle)
	configs, err := gen.GenerateBatch(200)
	require.NoError(t, err)

	for _, cfg := range configs {
		assert.True(t, netutil.IsRFC1918(cfg.IPNetwork),
			"network %s should be RFC 1918", cfg.IPNetwork)
	}
}

func TestAllVlanIDsInRange(t *testing.T) {
	t.Parallel()

	gen := generator.NewVlanGenerator(int64Ptr(42), generator.WanSingle)
	configs, err := gen.GenerateBatch(200)
	require.NoError(t, err)

	for _, cfg := range configs {
		assert.GreaterOrEqual(t, cfg.VlanID, uint16(generator.MinVlanID))
		assert.LessOrEqual(t, cfg.VlanID, uint16(generator.MaxVlanID))
	}
}

func TestDeterministicOutput(t *testing.T) {
	t.Parallel()

	gen1 := generator.NewVlanGenerator(int64Ptr(12345), generator.WanSingle)
	gen2 := generator.NewVlanGenerator(int64Ptr(12345), generator.WanSingle)

	configs1, err := gen1.GenerateBatch(10)
	require.NoError(t, err)

	configs2, err := gen2.GenerateBatch(10)
	require.NoError(t, err)

	for i := range configs1 {
		assert.Equal(t, configs1[i].VlanID, configs2[i].VlanID, "VLAN ID mismatch at index %d", i)
		assert.Equal(t, configs1[i].IPNetwork, configs2[i].IPNetwork, "network mismatch at index %d", i)
		assert.Equal(t, configs1[i].Description, configs2[i].Description, "description mismatch at index %d", i)
		assert.Equal(t, configs1[i].WanAssignment, configs2[i].WanAssignment, "WAN mismatch at index %d", i)
	}
}

func TestWanAssignmentSingle(t *testing.T) {
	t.Parallel()

	gen := generator.NewVlanGenerator(int64Ptr(42), generator.WanSingle)
	configs, err := gen.GenerateBatch(10)
	require.NoError(t, err)

	for _, cfg := range configs {
		assert.Equal(t, uint8(1), cfg.WanAssignment)
	}
}

func TestWanAssignmentMulti(t *testing.T) {
	t.Parallel()

	gen := generator.NewVlanGenerator(int64Ptr(42), generator.WanMulti)
	configs, err := gen.GenerateBatch(9)
	require.NoError(t, err)

	for i, cfg := range configs {
		expected := uint8(i%3) + 1
		assert.Equal(t, expected, cfg.WanAssignment, "WAN assignment at index %d", i)
	}
}

func TestWanAssignmentBalanced(t *testing.T) {
	t.Parallel()

	gen := generator.NewVlanGenerator(int64Ptr(42), generator.WanBalanced)
	configs, err := gen.GenerateBatch(100)
	require.NoError(t, err)

	counts := make(map[uint8]int)
	for _, cfg := range configs {
		assert.GreaterOrEqual(t, cfg.WanAssignment, uint8(1))
		assert.LessOrEqual(t, cfg.WanAssignment, uint8(3))
		counts[cfg.WanAssignment]++
	}

	for wan := uint8(1); wan <= 3; wan++ {
		assert.Positive(t, counts[wan], "WAN %d should have at least one assignment", wan)
	}
}

func TestMaxVlanGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping max VLAN generation test in short mode")
	}
	t.Parallel()

	gen := generator.NewVlanGenerator(int64Ptr(42), generator.WanSingle)
	configs, err := gen.GenerateBatch(generator.MaxUniqueVlans)
	require.NoError(t, err)
	assert.Len(t, configs, generator.MaxUniqueVlans)
}

func TestDescriptionFormat(t *testing.T) {
	t.Parallel()

	gen := generator.NewVlanGenerator(int64Ptr(42), generator.WanSingle)
	cfg, err := gen.GenerateSingle()
	require.NoError(t, err)

	assert.Contains(t, cfg.Description, "VLAN")
	assert.NotEmpty(t, cfg.Department)
}

func TestUsedCountTracking(t *testing.T) {
	t.Parallel()

	gen := generator.NewVlanGenerator(int64Ptr(42), generator.WanSingle)
	assert.Equal(t, 0, gen.UsedVlanIDCount())
	assert.Equal(t, 0, gen.UsedNetworkCount())

	_, err := gen.GenerateBatch(5)
	require.NoError(t, err)

	assert.Equal(t, 5, gen.UsedVlanIDCount())
	assert.Equal(t, 5, gen.UsedNetworkCount())
}

func TestParseWanAssignment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected generator.WanAssignment
		wantErr  bool
	}{
		{"single", "single", generator.WanSingle, false},
		{"multi", "multi", generator.WanMulti, false},
		{"balanced", "balanced", generator.WanBalanced, false},
		{"invalid", "roundrobin", generator.WanSingle, true},
		{"empty", "", generator.WanSingle, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := generator.ParseWanAssignment(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "must be single, multi, or balanced")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}
