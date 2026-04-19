package faker

import (
	"net/netip"
	"testing"

	"github.com/EvilBit-Labs/opnDossier/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFakeDHCPScopesOnePerStaticInterface(t *testing.T) {
	t.Parallel()

	interfaces := []model.Interface{
		{Name: "wan", Type: "dhcp"},
		{Name: "lan", Type: "static", IPAddress: "192.168.1.1", Subnet: "24"},
		{Name: "opt1", Type: "static", IPAddress: "10.0.0.1", Subnet: "24", Virtual: true},
	}

	scopes, err := fakeDHCPScopes(interfaces)
	require.NoError(t, err)

	require.Len(t, scopes, 2, "WAN excluded; LAN + opt1 each produce one scope")
	names := []string{scopes[0].Interface, scopes[1].Interface}
	assert.Contains(t, names, "lan")
	assert.Contains(t, names, "opt1")

	for _, s := range scopes {
		assert.True(t, s.Enabled)
		assert.NotEmpty(t, s.Range.From)
		assert.NotEmpty(t, s.Range.To)
		from, err := netip.ParseAddr(s.Range.From)
		require.NoError(t, err)
		to, err := netip.ParseAddr(s.Range.To)
		require.NoError(t, err)
		assert.True(t, from.Less(to), "DHCP range.from must be < range.to")
	}
}

func TestFakeDHCPScopesSkipsWhenFieldsMissing(t *testing.T) {
	t.Parallel()

	scopes, err := fakeDHCPScopes([]model.Interface{
		{Name: "lan", Type: "static", IPAddress: "", Subnet: "24"},
		{Name: "opt1", Type: "static", IPAddress: "10.0.0.1", Subnet: ""},
		{Name: "opt2", Type: "none"},
	})
	require.NoError(t, err)
	assert.Empty(t, scopes, "interfaces with missing fields produce no scope")
}

func TestFakeDHCPScopesErrorOnUnparseablePrefix(t *testing.T) {
	t.Parallel()

	_, err := fakeDHCPScopes([]model.Interface{
		{Name: "bad", Type: "static", IPAddress: "not-an-ip", Subnet: "24"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unparseable prefix")
	assert.Contains(t, err.Error(), "bad")
}
