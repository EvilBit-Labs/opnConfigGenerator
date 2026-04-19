// Tests in this file verify that opnConfigGenerator can act as an external
// consumer of opnDossier's public pkg/ API surface. They exercise the full
// reverse-serializer pipeline:
//
//	faker.NewCommonDevice -> serializer.Serialize -> MarshalConfig -> ParseConfig -> ConvertDocument
//
// These tests intentionally use github.com/EvilBit-Labs/opnDossier/pkg/model
// and github.com/EvilBit-Labs/opnDossier/pkg/parser/opnsense so that a build
// failure here signals a regression in the opnDossier public API contract.
package opnsensegen_test

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

// TestCommonDeviceRoundTripViaSerializer exercises the full reverse pipeline
// against a faker-generated device and asserts round-trip parity on the
// fields Phase 1 covers. Zero ConversionWarnings is the primary gate.
func TestCommonDeviceRoundTripViaSerializer(t *testing.T) {
	t.Parallel()

	original := faker.NewCommonDevice(
		faker.WithSeed(146),
		faker.WithVLANCount(2),
	)

	doc, err := serializer.Serialize(original)
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, opnsensegen.MarshalConfig(doc, &buf))

	parsed, err := opnsensegen.ParseConfig(buf.Bytes())
	require.NoError(t, err)

	device, warnings, err := opnsenseparser.ConvertDocument(parsed)
	require.NoError(t, err, "ConvertDocument must accept a serializer-produced document")
	require.NotNil(t, device)

	assert.Emptyf(t, warnings,
		"serializer output produced %d ConversionWarning(s): %+v", len(warnings), warnings)

	assert.Equal(t, model.DeviceTypeOPNsense, device.DeviceType)
	assert.Equal(t, original.System.Hostname, device.System.Hostname)
	assert.Equal(t, original.System.Domain, device.System.Domain)
	assert.Len(t, device.VLANs, 2)
	assert.NotEmpty(t, device.Interfaces)
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
	assert.Emptyf(t, warnings,
		"minimal config produced %d ConversionWarning(s): %+v", len(warnings), warnings)

	assert.Equal(t, model.DeviceTypeOPNsense, device.DeviceType)
	assert.Equal(t, "opnsense", device.System.Hostname)
}

// TestCommonDeviceNilDocument pins the consumer-visible error contract for
// ConvertDocument against nil input. Guards against a silent change where
// nil input starts returning a different error or panicking.
func TestCommonDeviceNilDocument(t *testing.T) {
	t.Parallel()

	device, warnings, err := opnsenseparser.ConvertDocument(nil)
	require.ErrorIs(t, err, opnsenseparser.ErrNilDocument,
		"nil document must return opnsenseparser.ErrNilDocument sentinel")
	assert.Nil(t, device)
	assert.Empty(t, warnings)
}
