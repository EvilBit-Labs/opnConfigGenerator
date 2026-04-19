package opnsense_test

import (
	"testing"

	serializer "github.com/EvilBit-Labs/opnConfigGenerator/internal/serializer/opnsense"
	"github.com/EvilBit-Labs/opnDossier/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSerializeDHCPEnabled(t *testing.T) {
	t.Parallel()

	in := []model.DHCPScope{{
		Interface: "lan",
		Enabled:   true,
		Range:     model.DHCPRange{From: "192.168.1.100", To: "192.168.1.200"},
		Gateway:   "192.168.1.1",
		DNSServer: "192.168.1.1",
	}}

	out := serializer.SerializeDHCP(in)

	require.NotNil(t, out.Items)
	lan, ok := out.Items["lan"]
	require.True(t, ok)
	assert.Equal(t, "1", lan.Enable)
	assert.Equal(t, "192.168.1.100", lan.Range.From)
	assert.Equal(t, "192.168.1.200", lan.Range.To)
	assert.Equal(t, "192.168.1.1", lan.Gateway)
	assert.Equal(t, "192.168.1.1", lan.Dnsserver)
}

func TestSerializeDHCPDisabled(t *testing.T) {
	t.Parallel()

	out := serializer.SerializeDHCP([]model.DHCPScope{{Interface: "lan", Enabled: false}})
	assert.Empty(t, out.Items["lan"].Enable)
}
