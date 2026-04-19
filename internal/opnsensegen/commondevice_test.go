// Tests in this file verify that opnConfigGenerator can act as an external
// consumer of opnDossier's public pkg/ API surface. Specifically, they exercise
// the file->CommonDevice pipeline described in NATS-146:
//
//	generate -> marshal XML -> parse XML -> ConvertDocument -> CommonDevice
//
// These tests intentionally use github.com/EvilBit-Labs/opnDossier/pkg/model
// and github.com/EvilBit-Labs/opnDossier/pkg/parser/opnsense so that a build
// failure here signals a regression in the opnDossier public API contract.
package opnsensegen_test

import (
	"bytes"
	"math/rand/v2"
	"net/netip"
	"testing"

	"github.com/EvilBit-Labs/opnConfigGenerator/internal/generator"
	"github.com/EvilBit-Labs/opnConfigGenerator/internal/opnsensegen"
	"github.com/EvilBit-Labs/opnDossier/pkg/model"
	opnsenseparser "github.com/EvilBit-Labs/opnDossier/pkg/parser/opnsense"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCommonDeviceRoundTrip exercises the full consumer pipeline against a
// realistic generated config: generate -> marshal XML -> parse XML -> convert.
// This is the primary acceptance test for NATS-146 acceptance criteria #1, #2,
// and #8 on the opnConfigGenerator side.
func TestCommonDeviceRoundTrip(t *testing.T) {
	t.Parallel()

	const (
		testHostname = "nats146-host"
		testDomain   = "test.local"
	)

	cfg, err := opnsensegen.LoadBaseConfig("../../testdata/base-config.xml")
	require.NoError(t, err)
	cfg.System.Hostname = testHostname
	cfg.System.Domain = testDomain

	vlans := []generator.VlanConfig{
		{
			VlanID:        42,
			IPNetwork:     netip.MustParsePrefix("10.42.7.0/24"),
			Description:   "IT VLAN 42",
			WanAssignment: 1,
			Department:    generator.DeptIT,
		},
		{
			VlanID:        100,
			IPNetwork:     netip.MustParsePrefix("10.100.0.0/24"),
			Description:   "Sales VLAN 100",
			WanAssignment: 2,
			Department:    generator.DeptSales,
		},
	}
	opnsensegen.InjectVlans(cfg, vlans, 6)

	//nolint:gosec // Deterministic fake data generation, not security-sensitive
	rng := rand.New(rand.NewPCG(42, 0))
	dhcpConfigs := []generator.DhcpServerConfig{
		generator.DeriveDHCPConfig(vlans[0], rng),
	}
	opnsensegen.InjectDHCP(cfg, dhcpConfigs, 6)

	// Round-trip through XML bytes to prove the consumer pipeline works against
	// on-disk representation, not just in-memory struct passing.
	var buf bytes.Buffer
	require.NoError(t, opnsensegen.MarshalConfig(cfg, &buf))

	parsed, err := opnsensegen.ParseConfig(buf.Bytes())
	require.NoError(t, err)

	device, warnings, err := opnsenseparser.ConvertDocument(parsed)
	require.NoError(t, err, "ConvertDocument must accept a generator-produced document")
	require.NotNil(t, device, "ConvertDocument must return a non-nil CommonDevice")

	// Zero warnings expected for output we ship. If warnings appear, surface
	// them in the failure so future regressions are diagnosable.
	assert.Empty(t, warnings,
		"generator output produced %d ConversionWarning(s): %+v", len(warnings), warnings)

	assert.Equal(t, model.DeviceTypeOPNsense, device.DeviceType,
		"DeviceType must be OPNsense for opnsense converter output")
	assert.Equal(t, testHostname, device.System.Hostname)
	assert.Equal(t, testDomain, device.System.Domain)
	assert.Len(t, device.VLANs, 2, "both injected VLANs must be present in CommonDevice")
	assert.NotEmpty(t, device.Interfaces, "injected VLAN interfaces must surface in CommonDevice")
}

// TestCommonDeviceMinimalConfig verifies ConvertDocument accepts the sparse
// base-config.xml fixture without error. This locks in the "minimum viable
// consumer input" contract: any valid OpnSenseDocument, even one with empty
// collections, must convert cleanly.
func TestCommonDeviceMinimalConfig(t *testing.T) {
	t.Parallel()

	cfg, err := opnsensegen.LoadBaseConfig("../../testdata/base-config.xml")
	require.NoError(t, err)

	device, warnings, err := opnsenseparser.ConvertDocument(cfg)
	require.NoError(t, err)
	require.NotNil(t, device)
	assert.Empty(t, warnings,
		"minimal config produced %d ConversionWarning(s): %+v", len(warnings), warnings)

	assert.Equal(t, model.DeviceTypeOPNsense, device.DeviceType)
	assert.Equal(t, "opnsense", device.System.Hostname)
}

// TestCommonDeviceNilDocument pins the consumer-visible error contract for
// ConvertDocument. Guards against a silent change where nil input starts
// returning a different error or panicking.
func TestCommonDeviceNilDocument(t *testing.T) {
	t.Parallel()

	device, warnings, err := opnsenseparser.ConvertDocument(nil)
	require.ErrorIs(t, err, opnsenseparser.ErrNilDocument,
		"nil document must return opnsenseparser.ErrNilDocument sentinel")
	assert.Nil(t, device)
	assert.Empty(t, warnings)
}
