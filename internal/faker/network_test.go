package faker

import (
	"net/netip"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFakeNetworkWANAndLANOnly(t *testing.T) {
	t.Parallel()

	rng, f := newRand(100)
	result, err := fakeNetwork(rng, f, 0)
	require.NoError(t, err)

	require.Len(t, result.Interfaces, 2, "WAN + LAN only when vlanCount == 0")
	names := []string{result.Interfaces[0].Name, result.Interfaces[1].Name}
	assert.Contains(t, names, "wan")
	assert.Contains(t, names, "lan")
	assert.Empty(t, result.VLANs)
}

func TestFakeNetworkProducesRequestedVLANs(t *testing.T) {
	t.Parallel()

	rng, f := newRand(101)
	result, err := fakeNetwork(rng, f, 5)
	require.NoError(t, err)

	assert.Len(t, result.VLANs, 5)
	assert.Len(t, result.Interfaces, 2+5, "WAN + LAN + one opt per VLAN")

	seenTags := map[string]bool{}
	for _, v := range result.VLANs {
		tag, err := strconv.Atoi(v.Tag)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, tag, vlanTagMin)
		assert.LessOrEqual(t, tag, vlanTagMax)
		assert.Falsef(t, seenTags[v.Tag], "VLAN tag %s duplicated", v.Tag)
		seenTags[v.Tag] = true
	}
}

func TestFakeNetworkLANIsRFC1918(t *testing.T) {
	t.Parallel()

	rng, f := newRand(102)
	result, err := fakeNetwork(rng, f, 0)
	require.NoError(t, err)

	for i := range result.Interfaces {
		if result.Interfaces[i].Name != "lan" {
			continue
		}
		addr, err := netip.ParseAddr(result.Interfaces[i].IPAddress)
		require.NoError(t, err)
		assert.Truef(t, addr.IsPrivate(), "LAN must be RFC 1918, got %s", addr)
		assert.Equal(t, "24", result.Interfaces[i].Subnet)
		return
	}
	t.Fatal("no lan interface in result")
}

func TestFakeNetworkDeterministic(t *testing.T) {
	t.Parallel()

	rngA, fa := newRand(42)
	rngB, fb := newRand(42)
	a, errA := fakeNetwork(rngA, fa, 3)
	require.NoError(t, errA)
	b, errB := fakeNetwork(rngB, fb, 3)
	require.NoError(t, errB)
	assert.Equal(t, a, b, "same seed must produce identical networkResult")
}

func TestPickUniqueTagReturnsErrorOnExhaustion(t *testing.T) {
	t.Parallel()

	// Pre-populate the used set with every valid tag.
	used := make(map[uint16]bool, vlanTagMax-vlanTagMin+1)
	for tag := uint16(vlanTagMin); tag <= vlanTagMax; tag++ {
		used[tag] = true
	}
	rng, _ := newRand(1)
	_, err := pickUniqueTag(rng, used)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exhausted")
}
