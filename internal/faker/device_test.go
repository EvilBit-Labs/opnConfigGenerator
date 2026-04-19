package faker

import (
	"testing"

	"github.com/EvilBit-Labs/opnDossier/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCommonDeviceDefaults(t *testing.T) {
	t.Parallel()

	dev, err := NewCommonDevice()
	require.NoError(t, err)
	require.NotNil(t, dev)
	assert.Equal(t, model.DeviceTypeOPNsense, dev.DeviceType)
	assert.NotEmpty(t, dev.System.Hostname)
	assert.NotEmpty(t, dev.System.Domain)
	assert.Len(t, dev.Interfaces, 2, "default shape: WAN + LAN, no VLANs")
	assert.Empty(t, dev.VLANs)
	assert.Empty(t, dev.FirewallRules)
	// LAN is the only static interface → one DHCP scope.
	assert.Len(t, dev.DHCP, 1)
}

func TestNewCommonDeviceDeterministic(t *testing.T) {
	t.Parallel()

	a, err := NewCommonDevice(WithSeed(99), WithVLANCount(3))
	require.NoError(t, err)
	b, err := NewCommonDevice(WithSeed(99), WithVLANCount(3))
	require.NoError(t, err)
	assert.Equal(t, a, b, "same seed + options must produce equal devices")
}

func TestNewCommonDeviceVLANCount(t *testing.T) {
	t.Parallel()

	dev, err := NewCommonDevice(WithSeed(1), WithVLANCount(4))
	require.NoError(t, err)
	assert.Len(t, dev.VLANs, 4)
	assert.Len(t, dev.Interfaces, 2+4, "WAN + LAN + 4 opt interfaces")
	assert.Len(t, dev.DHCP, 5, "LAN + 4 opts each carry a DHCP scope")
}

func TestNewCommonDeviceFirewallRulesOptIn(t *testing.T) {
	t.Parallel()

	without, err := NewCommonDevice(WithSeed(1))
	require.NoError(t, err)
	assert.Empty(t, without.FirewallRules)

	with, err := NewCommonDevice(WithSeed(1), WithFirewallRules(true))
	require.NoError(t, err)
	assert.NotEmpty(t, with.FirewallRules)
}

func TestNewCommonDeviceHostnameAndDomainOverride(t *testing.T) {
	t.Parallel()

	dev, err := NewCommonDevice(WithSeed(1), WithHostname("gateway"), WithDomain("example.test"))
	require.NoError(t, err)
	assert.Equal(t, "gateway", dev.System.Hostname)
	assert.Equal(t, "example.test", dev.System.Domain)
}

// TestNewCommonDeviceFuzzSeeds exercises the faker across many distinct seeds
// to catch any regression that would silently produce empty output or trigger
// the pickUnique* exhaustion paths under adversarial seed streams.
func TestNewCommonDeviceFuzzSeeds(t *testing.T) {
	t.Parallel()

	for seed := int64(1); seed <= 200; seed++ {
		dev, err := NewCommonDevice(WithSeed(seed), WithVLANCount(8))
		require.NoErrorf(t, err, "seed %d produced error", seed)
		require.NotNilf(t, dev, "seed %d produced nil device", seed)
		assert.Lenf(t, dev.VLANs, 8, "seed %d wrong VLAN count", seed)
		assert.Lenf(t, dev.Interfaces, 10, "seed %d wrong interface count (WAN+LAN+8 opts)", seed)
	}
}
