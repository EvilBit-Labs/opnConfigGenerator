package csvio_test

import (
	"bytes"
	"net/netip"
	"strings"
	"testing"

	"github.com/EvilBit-Labs/opnConfigGenerator/internal/csvio"
	"github.com/EvilBit-Labs/opnConfigGenerator/internal/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteAndReadRoundTrip(t *testing.T) {
	t.Parallel()

	vlans := []generator.VlanConfig{
		{VlanID: 42, IPNetwork: netip.MustParsePrefix("10.42.0.0/24"), Description: "IT VLAN 42", WanAssignment: 1},
		{
			VlanID:        100,
			IPNetwork:     netip.MustParsePrefix("10.100.0.0/24"),
			Description:   "Sales VLAN 100",
			WanAssignment: 2,
		},
		{
			VlanID:        200,
			IPNetwork:     netip.MustParsePrefix("172.16.5.0/24"),
			Description:   "Engineering VLAN 200",
			WanAssignment: 3,
		},
	}

	var buf bytes.Buffer
	err := csvio.WriteVlanCSV(&buf, vlans)
	require.NoError(t, err)

	result, err := csvio.ReadVlanCSV(&buf)
	require.NoError(t, err)

	assert.Len(t, result, 3)
	for i, v := range result {
		assert.Equal(t, vlans[i].VlanID, v.VlanID, "VLAN ID mismatch at %d", i)
		assert.Equal(t, vlans[i].IPNetwork, v.IPNetwork, "network mismatch at %d", i)
		assert.Equal(t, vlans[i].Description, v.Description, "description mismatch at %d", i)
		assert.Equal(t, vlans[i].WanAssignment, v.WanAssignment, "WAN mismatch at %d", i)
	}
}

func TestWriteCSVGermanHeaders(t *testing.T) {
	t.Parallel()

	vlans := []generator.VlanConfig{
		{VlanID: 42, IPNetwork: netip.MustParsePrefix("10.1.1.0/24"), Description: "test", WanAssignment: 1},
	}

	var buf bytes.Buffer
	err := csvio.WriteVlanCSV(&buf, vlans)
	require.NoError(t, err)

	content := buf.String()
	// Skip BOM (3 bytes).
	if len(content) > 3 {
		content = content[3:]
	}

	assert.True(t, strings.HasPrefix(content, "VLAN,IP Range,Beschreibung,WAN\n"),
		"should start with German headers, got: %q", content[:50])
}

func TestWriteCSVUTF8BOM(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := csvio.WriteVlanCSV(&buf, nil)
	require.NoError(t, err)

	data := buf.Bytes()
	assert.Equal(t, byte(0xEF), data[0])
	assert.Equal(t, byte(0xBB), data[1])
	assert.Equal(t, byte(0xBF), data[2])
}

func TestReadCSVInvalidVlanID(t *testing.T) {
	t.Parallel()

	input := "VLAN,IP Range,Beschreibung,WAN\n5,10.1.1.0/24,test,1\n"
	_, err := csvio.ReadVlanCSV(strings.NewReader(input))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "VLAN ID")
}

func TestReadCSVInvalidNetwork(t *testing.T) {
	t.Parallel()

	input := "VLAN,IP Range,Beschreibung,WAN\n100,invalid,test,1\n"
	_, err := csvio.ReadVlanCSV(strings.NewReader(input))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid network")
}

func TestReadCSVEmptyDescription(t *testing.T) {
	t.Parallel()

	input := "VLAN,IP Range,Beschreibung,WAN\n100,10.1.1.0/24,,1\n"
	_, err := csvio.ReadVlanCSV(strings.NewReader(input))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "description")
}

func TestReadCSVInvalidWAN(t *testing.T) {
	t.Parallel()

	input := "VLAN,IP Range,Beschreibung,WAN\n100,10.1.1.0/24,test,5\n"
	_, err := csvio.ReadVlanCSV(strings.NewReader(input))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "WAN assignment")
}

func TestReadCSVWrongHeaders(t *testing.T) {
	t.Parallel()

	input := "ID,Network,Name,WAN\n100,10.1.1.0/24,test,1\n"
	_, err := csvio.ReadVlanCSV(strings.NewReader(input))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "header")
}

func TestReadCSVEmpty(t *testing.T) {
	t.Parallel()

	input := "VLAN,IP Range,Beschreibung,WAN\n"
	vlans, err := csvio.ReadVlanCSV(strings.NewReader(input))
	require.NoError(t, err)
	assert.Empty(t, vlans)
}

func TestWriteCSVEmptySlice(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := csvio.WriteVlanCSV(&buf, []generator.VlanConfig{})
	require.NoError(t, err)

	// Should still have BOM + header.
	result, err := csvio.ReadVlanCSV(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)
	assert.Empty(t, result)
}
