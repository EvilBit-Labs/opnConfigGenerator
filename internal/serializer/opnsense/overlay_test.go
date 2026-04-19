package opnsense_test

import (
	"testing"

	"github.com/EvilBit-Labs/opnConfigGenerator/internal/faker"
	serializer "github.com/EvilBit-Labs/opnConfigGenerator/internal/serializer/opnsense"
	opnschema "github.com/EvilBit-Labs/opnDossier/pkg/schema/opnsense"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOverlayPreservesBaseConfigUnrelatedFields(t *testing.T) {
	t.Parallel()

	base := &opnschema.OpnSenseDocument{
		Version: "1.2.3",
		Theme:   "opnsense",
		System:  opnschema.System{Hostname: "base-host", Domain: "base.test"},
		Sysctl: []opnschema.SysctlItem{
			{Tunable: "net.inet.ip.forwarding", Value: "1", Descr: "route"},
		},
		CAs: []opnschema.CertificateAuthority{
			{Refid: "ca-1", Descr: "internal root"},
		},
	}
	device, err := faker.NewCommonDevice(faker.WithSeed(1), faker.WithVLANCount(2))
	require.NoError(t, err)

	merged, err := serializer.Overlay(base, device)
	require.NoError(t, err)

	// Fields outside Phase 1 scope must survive wholesale.
	assert.Equal(t, "opnsense", merged.Theme, "Theme must survive overlay")
	assert.Equal(t, "1.2.3", merged.Version, "Version must survive overlay")
	require.Len(t, merged.Sysctl, 1, "Sysctl must survive overlay")
	assert.Equal(t, "net.inet.ip.forwarding", merged.Sysctl[0].Tunable)
	require.Len(t, merged.CAs, 1, "CAs must survive overlay")
	assert.Equal(t, "ca-1", merged.CAs[0].Refid)

	// Phase 1 subsystems come from the device.
	assert.Equal(t, device.System.Hostname, merged.System.Hostname, "System replaced from device")
	assert.Len(t, merged.VLANs.VLAN, 2)
}

// TestOverlayReplacesDhcpdAndFilter seeds both Dhcpd and Filter in the base
// fixture and enables faker firewall rules, then asserts that overlay drops
// the base's Dhcpd/Filter content and installs the device's. Pins the
// wholesale-replace contract for both sections so a future silent shift to
// merge semantics fails this test.
func TestOverlayReplacesDhcpdAndFilter(t *testing.T) {
	t.Parallel()

	base := &opnschema.OpnSenseDocument{
		Version: "1.0",
		Dhcpd: opnschema.Dhcpd{
			Items: map[string]opnschema.DhcpdInterface{
				"lan": {
					Enable:  "1",
					Gateway: "10.99.99.1",
					Range: opnschema.Range{
						From: "10.99.99.10",
						To:   "10.99.99.20",
					},
				},
			},
		},
		Filter: opnschema.Filter{
			Rule: []opnschema.Rule{
				{Type: "block", Descr: "from base — must be dropped"},
			},
		},
	}
	device, err := faker.NewCommonDevice(
		faker.WithSeed(1),
		faker.WithVLANCount(2),
		faker.WithFirewallRules(true),
	)
	require.NoError(t, err)
	require.NotEmpty(t, device.FirewallRules)
	require.NotEmpty(t, device.DHCP)

	merged, err := serializer.Overlay(base, device)
	require.NoError(t, err)

	// Dhcpd — base's lan gateway 10.99.99.1 must NOT survive; the device's
	// LAN gateway replaces it.
	lan, ok := merged.Dhcpd.Items["lan"]
	require.True(t, ok, "merged document must have a lan dhcpd entry from the device")
	assert.NotEqual(t, "10.99.99.1", lan.Gateway,
		"base Dhcpd must be replaced wholesale; base gateway must not survive")

	// Filter — the base's block rule must not survive.
	for _, r := range merged.Filter.Rule {
		assert.NotEqual(t, "from base — must be dropped", r.Descr,
			"base Filter.Rule entries must be replaced wholesale")
	}
	assert.NotEmpty(t, merged.Filter.Rule, "device firewall rules must land in merged Filter")
}

func TestOverlayNilBase(t *testing.T) {
	t.Parallel()

	dev, err := faker.NewCommonDevice(faker.WithSeed(1))
	require.NoError(t, err)
	_, err = serializer.Overlay(nil, dev)
	require.ErrorIs(t, err, serializer.ErrNilBase)
}

func TestOverlayNilDeviceSurfacesSerializeError(t *testing.T) {
	t.Parallel()

	base := &opnschema.OpnSenseDocument{Version: "1.0"}
	_, err := serializer.Overlay(base, nil)
	require.ErrorIs(t, err, serializer.ErrNilDevice)
}
