package opnsense_test

import (
	"testing"

	serializer "github.com/EvilBit-Labs/opnConfigGenerator/internal/serializer/opnsense"
	"github.com/EvilBit-Labs/opnDossier/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestSerializeVLANs(t *testing.T) {
	t.Parallel()

	in := []model.VLAN{
		{VLANIf: "vlan0.42", PhysicalIf: "igb0", Tag: "42", Description: "IT"},
		{VLANIf: "vlan0.100", PhysicalIf: "igb0", Tag: "100", Description: "Sales"},
	}

	out := serializer.SerializeVLANs(in)

	assert.Len(t, out.VLAN, 2)
	assert.Equal(t, "42", out.VLAN[0].Tag)
	assert.Equal(t, "igb0", out.VLAN[0].If)
	assert.Equal(t, "vlan0.42", out.VLAN[0].Vlanif)
	assert.Equal(t, "IT", out.VLAN[0].Descr)
	// Spot-check the second entry so a regression that only populates the
	// first slice element doesn't pass.
	assert.Equal(t, "100", out.VLAN[1].Tag)
	assert.Equal(t, "vlan0.100", out.VLAN[1].Vlanif)
	assert.Equal(t, "Sales", out.VLAN[1].Descr)
}

func TestSerializeVLANsEmpty(t *testing.T) {
	t.Parallel()

	out := serializer.SerializeVLANs(nil)
	assert.Empty(t, out.VLAN)
}
