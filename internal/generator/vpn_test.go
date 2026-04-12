package generator_test

import (
	"testing"

	"github.com/EvilBit-Labs/opnConfigGenerator/internal/generator"
	"github.com/EvilBit-Labs/opnConfigGenerator/internal/netutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVpnGenerateConfigs(t *testing.T) {
	t.Parallel()

	gen := generator.NewVpnGenerator(seedPtr(42))
	configs, err := gen.GenerateConfigs(5)
	require.NoError(t, err)
	assert.Len(t, configs, 5)
}

func TestVpnZeroCountReturnsNil(t *testing.T) {
	t.Parallel()

	gen := generator.NewVpnGenerator(seedPtr(42))
	configs, err := gen.GenerateConfigs(0)
	require.NoError(t, err)
	assert.Nil(t, configs)
}

func TestVpnTunnelSubnetsUnique(t *testing.T) {
	t.Parallel()

	gen := generator.NewVpnGenerator(seedPtr(42))
	configs, err := gen.GenerateConfigs(10)
	require.NoError(t, err)

	seen := make(map[string]bool)
	for _, cfg := range configs {
		key := cfg.TunnelNetwork.String()
		assert.False(t, seen[key], "duplicate tunnel subnet: %s", key)
		seen[key] = true
	}
}

func TestVpnTunnelSubnetsAreRFC1918(t *testing.T) {
	t.Parallel()

	gen := generator.NewVpnGenerator(seedPtr(42))
	configs, err := gen.GenerateConfigs(10)
	require.NoError(t, err)

	for _, cfg := range configs {
		assert.True(t, netutil.IsRFC1918(cfg.TunnelNetwork),
			"tunnel %s should be RFC 1918", cfg.TunnelNetwork)
	}
}

func TestVpnOpenVPNConfig(t *testing.T) {
	t.Parallel()

	gen := generator.NewVpnGenerator(seedPtr(42))
	configs, err := gen.GenerateConfigsOfType(generator.VpnOpenVPN, 3)
	require.NoError(t, err)

	for _, cfg := range configs {
		assert.Equal(t, generator.VpnOpenVPN, cfg.Type)
		assert.Equal(t, "udp", cfg.Protocol)
		assert.Equal(t, "AES-256-GCM", cfg.Cipher)
		assert.True(t, cfg.TLSAuth)
		assert.GreaterOrEqual(t, cfg.Port, 1194)
		assert.Less(t, cfg.Port, 1294)
	}
}

func TestVpnWireGuardConfig(t *testing.T) {
	t.Parallel()

	gen := generator.NewVpnGenerator(seedPtr(42))
	configs, err := gen.GenerateConfigsOfType(generator.VpnWireGuard, 3)
	require.NoError(t, err)

	for _, cfg := range configs {
		assert.Equal(t, generator.VpnWireGuard, cfg.Type)
		assert.NotEmpty(t, cfg.PublicKey)
		assert.NotEmpty(t, cfg.ListenKey)
		assert.GreaterOrEqual(t, cfg.Port, 51820)
		assert.Less(t, cfg.Port, 51920)
	}
}

func TestVpnIPSecConfig(t *testing.T) {
	t.Parallel()

	gen := generator.NewVpnGenerator(seedPtr(42))
	configs, err := gen.GenerateConfigsOfType(generator.VpnIPSec, 3)
	require.NoError(t, err)

	for _, cfg := range configs {
		assert.Equal(t, generator.VpnIPSec, cfg.Type)
		assert.Equal(t, 2, cfg.IKEVersion)
		assert.GreaterOrEqual(t, cfg.DHGroup, 14)
		assert.Equal(t, "sha256", cfg.HashAlgo)
		assert.Equal(t, 500, cfg.Port)
	}
}

func TestVpnConfigsHaveDescriptions(t *testing.T) {
	t.Parallel()

	gen := generator.NewVpnGenerator(seedPtr(42))
	configs, err := gen.GenerateConfigs(10)
	require.NoError(t, err)

	for _, cfg := range configs {
		assert.NotEmpty(t, cfg.Description)
		assert.NotEmpty(t, cfg.Name)
		assert.NotEmpty(t, cfg.ID)
	}
}

func TestVpnConfigsHaveUniqueIDs(t *testing.T) {
	t.Parallel()

	gen := generator.NewVpnGenerator(seedPtr(42))
	configs, err := gen.GenerateConfigs(20)
	require.NoError(t, err)

	seen := make(map[string]bool)
	for _, cfg := range configs {
		assert.False(t, seen[cfg.ID], "duplicate VPN ID: %s", cfg.ID)
		seen[cfg.ID] = true
	}
}

func TestVpnSubnetExhaustion(t *testing.T) {
	t.Parallel()

	gen := generator.NewVpnGenerator(seedPtr(42))
	// 254 tunnel subnets available (10.200.1-254.0/24).
	_, err := gen.GenerateConfigs(254)
	require.NoError(t, err)

	// 255th should fail.
	_, err = gen.GenerateConfigs(1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exhausted")
}
