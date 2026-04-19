package opnsense_test

import (
	"testing"

	serializer "github.com/EvilBit-Labs/opnConfigGenerator/internal/serializer/opnsense"
	"github.com/EvilBit-Labs/opnDossier/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSerializeInterfacesKeyedByName(t *testing.T) {
	t.Parallel()

	in := []model.Interface{
		{Name: "wan", PhysicalIf: "igb1", Enabled: true, Type: "dhcp"},
		{Name: "lan", PhysicalIf: "igb0", Enabled: true, Type: "static", IPAddress: "192.168.1.1", Subnet: "24"},
		{
			Name:        "opt1",
			PhysicalIf:  "vlan0.42",
			Enabled:     true,
			Type:        "static",
			IPAddress:   "10.42.0.1",
			Subnet:      "24",
			Description: "IT",
		},
	}

	out := serializer.SerializeInterfaces(in)

	require.NotNil(t, out.Items)
	require.Len(t, out.Items, 3)

	wan, ok := out.Items["wan"]
	require.True(t, ok)
	assert.Equal(t, "1", wan.Enable)
	assert.Equal(t, "dhcp", wan.IPAddr)
	assert.Equal(t, "igb1", wan.If)

	lan, ok := out.Items["lan"]
	require.True(t, ok)
	assert.Equal(t, "192.168.1.1", lan.IPAddr)
	assert.Equal(t, "24", lan.Subnet)

	opt1, ok := out.Items["opt1"]
	require.True(t, ok)
	assert.Equal(t, "IT", opt1.Descr)
	assert.Equal(t, "vlan0.42", opt1.If)
}

func TestSerializeInterfacesDisabled(t *testing.T) {
	t.Parallel()

	out := serializer.SerializeInterfaces([]model.Interface{{Name: "lan", Enabled: false}})
	assert.Empty(t, out.Items["lan"].Enable)
}

func TestSerializeInterfacesTypeNoneLeavesAddressingEmpty(t *testing.T) {
	t.Parallel()

	out := serializer.SerializeInterfaces([]model.Interface{{
		Name: "opt1", Enabled: true, Type: "", IPAddress: "1.2.3.4", Subnet: "24",
	}})
	assert.Empty(t, out.Items["opt1"].IPAddr)
	assert.Empty(t, out.Items["opt1"].Subnet)
}
