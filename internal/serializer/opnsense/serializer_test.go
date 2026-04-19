package opnsense_test

import (
	"bytes"
	"testing"

	"github.com/EvilBit-Labs/opnConfigGenerator/internal/faker"
	"github.com/EvilBit-Labs/opnConfigGenerator/internal/opnsensegen"
	serializer "github.com/EvilBit-Labs/opnConfigGenerator/internal/serializer/opnsense"
	"github.com/EvilBit-Labs/opnDossier/pkg/model"
	opnsenseparser "github.com/EvilBit-Labs/opnDossier/pkg/parser/opnsense"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRoundTrip is the primary acceptance gate for R1. It exercises the
// full pipeline: faker → Serialize → MarshalConfig → ParseConfig →
// ConvertDocument, then asserts per-field parity on every field Phase 1
// claims to cover. Count-only assertions were insufficient — they missed
// silent field drops (e.g., Interface.Type, Interface.Virtual).
func TestRoundTrip(t *testing.T) {
	t.Parallel()

	original, err := faker.NewCommonDevice(
		faker.WithSeed(2026),
		faker.WithVLANCount(3),
		faker.WithFirewallRules(true),
	)
	require.NoError(t, err)

	doc, err := serializer.Serialize(original)
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, opnsensegen.MarshalConfig(doc, &buf))

	parsed, err := opnsensegen.ParseConfig(buf.Bytes())
	require.NoError(t, err)

	roundTripped, warnings, err := opnsenseparser.ConvertDocument(parsed)
	require.NoError(t, err)
	assert.Emptyf(t, warnings, "ConvertDocument warnings: %+v", warnings)

	require.NotNil(t, roundTripped)
	assert.Equal(t, model.DeviceTypeOPNsense, roundTripped.DeviceType)

	// System parity.
	assert.Equal(t, original.System.Hostname, roundTripped.System.Hostname)
	assert.Equal(t, original.System.Domain, roundTripped.System.Domain)

	// Interface per-field parity, keyed by Name.
	require.Len(t, roundTripped.Interfaces, len(original.Interfaces))
	origIfaces := interfacesByName(original.Interfaces)
	rtIfaces := interfacesByName(roundTripped.Interfaces)
	for name, want := range origIfaces {
		got, ok := rtIfaces[name]
		require.Truef(t, ok, "interface %q missing on round-trip", name)
		assert.Equalf(t, want.Type, got.Type, "interface %q Type", name)
		assert.Equalf(t, want.Virtual, got.Virtual, "interface %q Virtual", name)
		assert.Equalf(t, want.Description, got.Description, "interface %q Description", name)
		if want.Type == "static" {
			assert.Equalf(t, want.IPAddress, got.IPAddress, "interface %q IPAddress", name)
			assert.Equalf(t, want.Subnet, got.Subnet, "interface %q Subnet", name)
		}
	}

	// VLAN per-field parity, keyed by Tag because order is not guaranteed.
	require.Len(t, roundTripped.VLANs, len(original.VLANs))
	origVLANs := vlansByTag(original.VLANs)
	rtVLANs := vlansByTag(roundTripped.VLANs)
	for tag, want := range origVLANs {
		got, ok := rtVLANs[tag]
		require.Truef(t, ok, "VLAN %q missing on round-trip", tag)
		assert.Equalf(t, want.VLANIf, got.VLANIf, "VLAN %q VLANIf", tag)
		assert.Equalf(t, want.PhysicalIf, got.PhysicalIf, "VLAN %q PhysicalIf", tag)
		assert.Equalf(t, want.Description, got.Description, "VLAN %q Description", tag)
	}

	// DHCP scope parity, keyed by interface name.
	require.Len(t, roundTripped.DHCP, len(original.DHCP))
	origDHCP := dhcpByInterface(original.DHCP)
	rtDHCP := dhcpByInterface(roundTripped.DHCP)
	for iface, want := range origDHCP {
		got, ok := rtDHCP[iface]
		require.Truef(t, ok, "DHCP scope for %q missing on round-trip", iface)
		assert.Equalf(t, want.Range.From, got.Range.From, "DHCP %q Range.From", iface)
		assert.Equalf(t, want.Range.To, got.Range.To, "DHCP %q Range.To", iface)
		assert.Equalf(t, want.Gateway, got.Gateway, "DHCP %q Gateway", iface)
		assert.Equalf(t, want.DNSServer, got.DNSServer, "DHCP %q DNSServer", iface)
	}

	// Firewall rule per-field parity. Faker emits rules keyed to an
	// interface name, so we compare by Interfaces[0] + Type to find pairs.
	require.Len(t, roundTripped.FirewallRules, len(original.FirewallRules))
	origRules := firewallRulesByInterface(original.FirewallRules)
	rtRules := firewallRulesByInterface(roundTripped.FirewallRules)
	for iface, want := range origRules {
		got, ok := rtRules[iface]
		require.Truef(t, ok, "firewall rule for interface %q missing on round-trip", iface)
		assert.Equalf(t, want.Type, got.Type, "rule %q Type", iface)
		assert.Equalf(t, want.Description, got.Description, "rule %q Description", iface)
		assert.Equalf(t, want.IPProtocol, got.IPProtocol, "rule %q IPProtocol", iface)
		assert.Equalf(t, want.Source.Address, got.Source.Address, "rule %q Source.Address", iface)
		assert.Equalf(t, want.Destination.Address, got.Destination.Address, "rule %q Destination.Address", iface)
	}
}

// TestRoundTripByteStable asserts MarshalConfig output is byte-identical
// across repeated calls on the same input. A single repeat is not enough
// to catch a regression in sortMapBackedSections because Go's map iteration
// is randomized per-encode; 10 iterations provide high confidence.
func TestRoundTripByteStable(t *testing.T) {
	t.Parallel()

	device, err := faker.NewCommonDevice(
		faker.WithSeed(99),
		faker.WithVLANCount(4),
		faker.WithFirewallRules(true),
	)
	require.NoError(t, err)
	doc, err := serializer.Serialize(device)
	require.NoError(t, err)

	var first bytes.Buffer
	require.NoError(t, opnsensegen.MarshalConfig(doc, &first))

	for i := range 10 {
		var next bytes.Buffer
		require.NoError(t, opnsensegen.MarshalConfig(doc, &next))
		require.Equalf(t, first.Bytes(), next.Bytes(), "marshal #%d diverged", i+1)
	}
}

func TestSerializeNilDevice(t *testing.T) {
	t.Parallel()

	_, err := serializer.Serialize(nil)
	require.ErrorIs(t, err, serializer.ErrNilDevice)
}

func interfacesByName(xs []model.Interface) map[string]model.Interface {
	m := make(map[string]model.Interface, len(xs))
	for _, x := range xs {
		m[x.Name] = x
	}
	return m
}

func vlansByTag(xs []model.VLAN) map[string]model.VLAN {
	m := make(map[string]model.VLAN, len(xs))
	for _, x := range xs {
		m[x.Tag] = x
	}
	return m
}

func dhcpByInterface(xs []model.DHCPScope) map[string]model.DHCPScope {
	m := make(map[string]model.DHCPScope, len(xs))
	for _, x := range xs {
		m[x.Interface] = x
	}
	return m
}

func firewallRulesByInterface(xs []model.FirewallRule) map[string]model.FirewallRule {
	m := make(map[string]model.FirewallRule, len(xs))
	for _, x := range xs {
		if len(x.Interfaces) == 0 {
			continue
		}
		m[x.Interfaces[0]] = x
	}
	return m
}
