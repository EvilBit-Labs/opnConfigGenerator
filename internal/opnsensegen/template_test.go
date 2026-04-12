package opnsensegen_test

import (
	"bytes"
	"math/rand/v2"
	"net/netip"
	"strings"
	"testing"

	"github.com/EvilBit-Labs/opnConfigGenerator/internal/generator"
	"github.com/EvilBit-Labs/opnConfigGenerator/internal/opnsensegen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadBaseConfig(t *testing.T) {
	t.Parallel()

	cfg, err := opnsensegen.LoadBaseConfig("../../testdata/base-config.xml")
	require.NoError(t, err)

	assert.Equal(t, "1.0", cfg.Version)
	assert.Equal(t, "opnsense", cfg.System.Hostname)
	assert.Equal(t, "localdomain", cfg.System.Domain)
}

func TestParseConfig(t *testing.T) {
	t.Parallel()

	xmlData := []byte(`<?xml version="1.0"?>
<opnsense>
  <version>1.0</version>
  <system>
    <hostname>test</hostname>
    <domain>test.local</domain>
  </system>
  <vlans/>
  <interfaces/>
  <dhcpd/>
  <filter/>
</opnsense>`)

	cfg, err := opnsensegen.ParseConfig(xmlData)
	require.NoError(t, err)
	assert.Equal(t, "test", cfg.System.Hostname)
}

func TestInjectVlans(t *testing.T) {
	t.Parallel()

	cfg, err := opnsensegen.LoadBaseConfig("../../testdata/base-config.xml")
	require.NoError(t, err)

	vlans := []generator.VlanConfig{
		{VlanID: 42, IPNetwork: netip.MustParsePrefix("10.42.7.0/24"), Description: "IT VLAN 42", WanAssignment: 1},
		{
			VlanID:        100,
			IPNetwork:     netip.MustParsePrefix("10.100.0.0/24"),
			Description:   "Sales VLAN 100",
			WanAssignment: 2,
		},
	}

	opnsensegen.InjectVlans(cfg, vlans, 6)

	assert.Len(t, cfg.VLANs.VLAN, 2)
	assert.Equal(t, "42", cfg.VLANs.VLAN[0].Tag)
	assert.Equal(t, "IT VLAN 42", cfg.VLANs.VLAN[0].Descr)

	// Verify interfaces were added to the map.
	assert.Len(t, cfg.Interfaces.Items, 2)
	opt6, ok := cfg.Interfaces.Items["opt6"]
	require.True(t, ok)
	assert.Equal(t, "10.42.7.1", opt6.IPAddr)
	assert.Equal(t, "IT VLAN 42", opt6.Descr)
}

func TestInjectFirewallRules(t *testing.T) {
	t.Parallel()

	cfg, err := opnsensegen.LoadBaseConfig("../../testdata/base-config.xml")
	require.NoError(t, err)

	rules := []generator.FirewallRule{
		{
			RuleID: "r1", Action: "pass", Protocol: "tcp", Direction: "in",
			Source: "10.1.1.0/24", Destination: "any", Ports: "80,443",
			Description: "Allow HTTP", Interface: "opt6", Tracker: 12345,
		},
		{
			RuleID: "r2", Action: "block", Protocol: "any", Direction: "in",
			Source: "any", Destination: "any", Ports: "any",
			Description: "Block all", Interface: "opt6", Tracker: 67890, Log: true,
		},
	}

	opnsensegen.InjectFirewallRules(cfg, rules)
	assert.Len(t, cfg.Filter.Rule, 2)
	assert.Equal(t, "pass", cfg.Filter.Rule[0].Type)
	assert.Equal(t, "Allow HTTP", cfg.Filter.Rule[0].Descr)

	// Verify "any" source path produces Source.Any field.
	assert.NotNil(t, cfg.Filter.Rule[1].Source.Any, "source 'any' should set Any field")
	assert.Empty(t, cfg.Filter.Rule[1].Source.Network, "source 'any' should not set Network")
}

func TestInjectDHCP(t *testing.T) {
	t.Parallel()

	cfg, err := opnsensegen.LoadBaseConfig("../../testdata/base-config.xml")
	require.NoError(t, err)

	//nolint:gosec // Deterministic fake data generation, not security-sensitive
	rng := rand.New(rand.NewPCG(42, 0))
	vlans := []generator.VlanConfig{
		{
			VlanID:        42,
			IPNetwork:     netip.MustParsePrefix("10.42.7.0/24"),
			Description:   "IT",
			WanAssignment: 1,
			Department:    generator.DeptIT,
		},
	}
	dhcpConfigs := []generator.DhcpServerConfig{
		generator.DeriveDHCPConfig(vlans[0], rng),
	}

	opnsensegen.InjectDHCP(cfg, dhcpConfigs, 6)

	// Verify DHCP was added to the map.
	assert.Len(t, cfg.Dhcpd.Items, 1)
	dhcpIface, ok := cfg.Dhcpd.Items["opt6"]
	require.True(t, ok)
	assert.Equal(t, "1", dhcpIface.Enable)
}

func TestMarshalRoundTrip(t *testing.T) {
	t.Parallel()

	cfg, err := opnsensegen.LoadBaseConfig("../../testdata/base-config.xml")
	require.NoError(t, err)

	var buf bytes.Buffer
	err = opnsensegen.MarshalConfig(cfg, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "<?xml")
	assert.Contains(t, output, "<opnsense>")
	assert.Contains(t, output, "<hostname>opnsense</hostname>")
	assert.Contains(t, output, "</opnsense>")
}

func TestMarshalWithInjectedData(t *testing.T) {
	t.Parallel()

	cfg, err := opnsensegen.LoadBaseConfig("../../testdata/base-config.xml")
	require.NoError(t, err)

	vlans := []generator.VlanConfig{
		{VlanID: 42, IPNetwork: netip.MustParsePrefix("10.42.7.0/24"), Description: "IT VLAN 42", WanAssignment: 1},
	}

	opnsensegen.InjectVlans(cfg, vlans, 6)

	var buf bytes.Buffer
	err = opnsensegen.MarshalConfig(cfg, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "<tag>42</tag>")
	assert.Contains(t, output, "IT VLAN 42")
	assert.Contains(t, output, "10.42.7.1")
}

func TestMarshalXMLEscaping(t *testing.T) {
	t.Parallel()

	cfg, err := opnsensegen.LoadBaseConfig("../../testdata/base-config.xml")
	require.NoError(t, err)

	vlans := []generator.VlanConfig{
		{
			VlanID:        42,
			IPNetwork:     netip.MustParsePrefix("10.42.7.0/24"),
			Description:   "Test & <Special> \"Chars\"",
			WanAssignment: 1,
		},
	}

	opnsensegen.InjectVlans(cfg, vlans, 6)

	var buf bytes.Buffer
	err = opnsensegen.MarshalConfig(cfg, &buf)
	require.NoError(t, err)

	output := buf.String()
	// XML encoder should escape special characters.
	assert.Contains(t, output, "&amp;")
	assert.Contains(t, output, "&lt;Special&gt;")
	assert.True(t, strings.Contains(output, "&#34;Chars&#34;") || strings.Contains(output, "&quot;Chars&quot;"),
		"quotes should be escaped")
}
