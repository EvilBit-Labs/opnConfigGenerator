package faker

import (
	"testing"

	"github.com/EvilBit-Labs/opnDossier/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCommonDeviceDefaults(t *testing.T) {
	t.Parallel()

	dev := NewCommonDevice()
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

	a := NewCommonDevice(WithSeed(99), WithVLANCount(3))
	b := NewCommonDevice(WithSeed(99), WithVLANCount(3))
	assert.Equal(t, a, b, "same seed + options must produce equal devices")
}

func TestNewCommonDeviceVLANCount(t *testing.T) {
	t.Parallel()

	dev := NewCommonDevice(WithSeed(1), WithVLANCount(4))
	assert.Len(t, dev.VLANs, 4)
	assert.Len(t, dev.Interfaces, 2+4, "WAN + LAN + 4 opt interfaces")
	assert.Len(t, dev.DHCP, 5, "LAN + 4 opts each carry a DHCP scope")
}

func TestNewCommonDeviceFirewallRulesOptIn(t *testing.T) {
	t.Parallel()

	without := NewCommonDevice(WithSeed(1))
	assert.Empty(t, without.FirewallRules)

	with := NewCommonDevice(WithSeed(1), WithFirewallRules(true))
	assert.NotEmpty(t, with.FirewallRules)
}

func TestNewCommonDeviceHostnameAndDomainOverride(t *testing.T) {
	t.Parallel()

	dev := NewCommonDevice(WithSeed(1), WithHostname("gateway"), WithDomain("example.test"))
	assert.Equal(t, "gateway", dev.System.Hostname)
	assert.Equal(t, "example.test", dev.System.Domain)
}
