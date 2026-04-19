package opnsense_test

import (
	"testing"

	serializer "github.com/EvilBit-Labs/opnConfigGenerator/internal/serializer/opnsense"
	"github.com/EvilBit-Labs/opnDossier/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestSerializeSystemPopulatedFields(t *testing.T) {
	t.Parallel()

	in := model.System{
		Hostname:    "gw",
		Domain:      "example.test",
		Timezone:    "UTC",
		Language:    "en_US",
		DNSServers:  []string{"1.1.1.1", "9.9.9.9"},
		TimeServers: []string{"0.pool.ntp.org"},
	}

	out := serializer.SerializeSystem(in)

	assert.Equal(t, "gw", out.Hostname)
	assert.Equal(t, "example.test", out.Domain)
	assert.Equal(t, "UTC", out.Timezone)
	assert.Equal(t, "en_US", out.Language)
	assert.Equal(t, "1.1.1.1 9.9.9.9", out.DNSServer)
	assert.Equal(t, "0.pool.ntp.org", out.TimeServers)
	assert.Equal(t, "https", out.WebGUI.Protocol)
	assert.Equal(t, "wheel", out.SSH.Group)
}

func TestSerializeSystemEmptyStillHasDefaults(t *testing.T) {
	t.Parallel()

	out := serializer.SerializeSystem(model.System{})
	assert.Equal(t, "https", out.WebGUI.Protocol)
	assert.Equal(t, "wheel", out.SSH.Group)
}
