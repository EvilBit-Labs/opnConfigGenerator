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
// ConvertDocument. The CommonDevice that comes out must match the one that
// went in on the fields Phase 1 covers.
func TestRoundTrip(t *testing.T) {
	t.Parallel()

	original := faker.NewCommonDevice(
		faker.WithSeed(2026),
		faker.WithVLANCount(3),
		faker.WithFirewallRules(true),
	)

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
	assert.Equal(t, original.System.Hostname, roundTripped.System.Hostname)
	assert.Equal(t, original.System.Domain, roundTripped.System.Domain)
	assert.Len(t, roundTripped.VLANs, len(original.VLANs))
	assert.Len(t, roundTripped.Interfaces, len(original.Interfaces))
	assert.Len(t, roundTripped.DHCP, len(original.DHCP))
	assert.Len(t, roundTripped.FirewallRules, len(original.FirewallRules))
}

func TestSerializeNilDevice(t *testing.T) {
	t.Parallel()

	_, err := serializer.Serialize(nil)
	require.ErrorIs(t, err, serializer.ErrNilDevice)
}
