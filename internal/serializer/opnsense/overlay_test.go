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
	}
	device := faker.NewCommonDevice(faker.WithSeed(1), faker.WithVLANCount(2))

	merged, err := serializer.Overlay(base, device)
	require.NoError(t, err)

	assert.Equal(t, "opnsense", merged.Theme, "Theme must survive overlay")
	assert.Equal(t, "1.2.3", merged.Version, "Version must survive overlay")
	assert.Equal(t, device.System.Hostname, merged.System.Hostname, "System replaced from device")
	assert.Len(t, merged.VLANs.VLAN, 2)
}

func TestOverlayNilBase(t *testing.T) {
	t.Parallel()

	_, err := serializer.Overlay(nil, faker.NewCommonDevice(faker.WithSeed(1)))
	require.ErrorIs(t, err, serializer.ErrNilBase)
}

func TestOverlayNilDeviceSurfacesSerializeError(t *testing.T) {
	t.Parallel()

	base := &opnschema.OpnSenseDocument{Version: "1.0"}
	_, err := serializer.Overlay(base, nil)
	require.ErrorIs(t, err, serializer.ErrNilDevice)
}
