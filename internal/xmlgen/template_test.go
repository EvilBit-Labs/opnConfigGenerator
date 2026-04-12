package xmlgen_test

import (
	"bytes"
	"math/rand/v2"
	"net/netip"
	"strings"
	"testing"

	"github.com/EvilBit-Labs/opnConfigGenerator/internal/generator"
	"github.com/EvilBit-Labs/opnConfigGenerator/internal/xmlgen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadBaseConfig(t *testing.T) {
	t.Parallel()

	cfg, err := xmlgen.LoadBaseConfig("../../testdata/base-config.xml")
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

	cfg, err := xmlgen.ParseConfig(xmlData)
	require.NoError(t, err)
	assert.Equal(t, "test", cfg.System.Hostname)
}

func TestInjectVlans(t *testing.T) {
	t.Parallel()

	cfg, err := xmlgen.LoadBaseConfig("../../testdata/base-config.xml")
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

	xmlgen.InjectVlans(cfg, vlans, 6)

	assert.Len(t, cfg.VLANs.VLANs, 2)
	assert.Equal(t, uint16(42), cfg.VLANs.VLANs[0].Tag)
	assert.Equal(t, "IT VLAN 42", cfg.VLANs.VLANs[0].Descr)
	assert.Len(t, cfg.Interfaces.Entries, 2)
	assert.Equal(t, "opt6", cfg.Interfaces.Entries[0].XMLName.Local)
	assert.Equal(t, "10.42.7.1", cfg.Interfaces.Entries[0].IPAddr)
}

func TestInjectFirewallRules(t *testing.T) {
	t.Parallel()

	cfg, err := xmlgen.LoadBaseConfig("../../testdata/base-config.xml")
	require.NoError(t, err)

	rules := []generator.FirewallRule{
		{
			RuleID: "r1", Action: "pass", Protocol: "tcp", Direction: "in",
			Source: "10.1.1.0/24", Destination: "any", Ports: "80,443",
			Description: "Allow HTTP", Interface: "opt6", Tracker: 12345,
		},
	}

	xmlgen.InjectFirewallRules(cfg, rules)
	assert.Len(t, cfg.Filter.Rules, 1)
	assert.Equal(t, "pass", cfg.Filter.Rules[0].Type)
	assert.Equal(t, "Allow HTTP", cfg.Filter.Rules[0].Descr)
}

func TestInjectDHCP(t *testing.T) {
	t.Parallel()

	cfg, err := xmlgen.LoadBaseConfig("../../testdata/base-config.xml")
	require.NoError(t, err)

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

	xmlgen.InjectDHCP(cfg, vlans, dhcpConfigs, 6)
	assert.Len(t, cfg.Dhcpd.Entries, 1)
}

func TestMarshalRoundTrip(t *testing.T) {
	t.Parallel()

	cfg, err := xmlgen.LoadBaseConfig("../../testdata/base-config.xml")
	require.NoError(t, err)

	var buf bytes.Buffer
	err = xmlgen.MarshalConfig(cfg, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "<?xml")
	assert.Contains(t, output, "<opnsense>")
	assert.Contains(t, output, "<hostname>opnsense</hostname>")
	assert.Contains(t, output, "</opnsense>")
}

func TestMarshalWithInjectedData(t *testing.T) {
	t.Parallel()

	cfg, err := xmlgen.LoadBaseConfig("../../testdata/base-config.xml")
	require.NoError(t, err)

	vlans := []generator.VlanConfig{
		{VlanID: 42, IPNetwork: netip.MustParsePrefix("10.42.7.0/24"), Description: "IT VLAN 42", WanAssignment: 1},
	}

	xmlgen.InjectVlans(cfg, vlans, 6)

	var buf bytes.Buffer
	err = xmlgen.MarshalConfig(cfg, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "<tag>42</tag>")
	assert.Contains(t, output, "IT VLAN 42")
	assert.Contains(t, output, "10.42.7.1")
}

func TestMarshalXMLEscaping(t *testing.T) {
	t.Parallel()

	cfg, err := xmlgen.LoadBaseConfig("../../testdata/base-config.xml")
	require.NoError(t, err)

	vlans := []generator.VlanConfig{
		{
			VlanID:        42,
			IPNetwork:     netip.MustParsePrefix("10.42.7.0/24"),
			Description:   "Test & <Special> \"Chars\"",
			WanAssignment: 1,
		},
	}

	xmlgen.InjectVlans(cfg, vlans, 6)

	var buf bytes.Buffer
	err = xmlgen.MarshalConfig(cfg, &buf)
	require.NoError(t, err)

	output := buf.String()
	// XML encoder should escape special characters.
	assert.Contains(t, output, "&amp;")
	assert.Contains(t, output, "&lt;Special&gt;")
	assert.True(t, strings.Contains(output, "&#34;Chars&#34;") || strings.Contains(output, "&quot;Chars&quot;"),
		"quotes should be escaped")
}
