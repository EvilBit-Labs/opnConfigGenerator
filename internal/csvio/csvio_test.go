package csvio_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/EvilBit-Labs/opnConfigGenerator/internal/csvio"
	"github.com/EvilBit-Labs/opnConfigGenerator/internal/faker"
	"github.com/EvilBit-Labs/opnDossier/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteVlanCSVFromCommonDevice(t *testing.T) {
	t.Parallel()

	dev := faker.NewCommonDevice(faker.WithSeed(7), faker.WithVLANCount(2))

	var buf bytes.Buffer
	require.NoError(t, csvio.WriteVlanCSV(&buf, dev))

	// Strip BOM.
	content := strings.TrimPrefix(buf.String(), "\ufeff")
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	require.Len(t, lines, 3, "header + 2 data rows")
	assert.Equal(t, "VLAN,IP Range,Beschreibung,WAN", lines[0])
}

func TestWriteVlanCSVNilDevice(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := csvio.WriteVlanCSV(&buf, nil)
	require.Error(t, err)
}

func TestWriteVlanCSVHeaders(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	require.NoError(t, csvio.WriteVlanCSV(&buf, &model.CommonDevice{}))

	content := buf.String()
	// BOM is 3 bytes.
	require.GreaterOrEqual(t, len(content), 3)
	assert.Equal(t, byte(0xEF), content[0])
	assert.Equal(t, byte(0xBB), content[1])
	assert.Equal(t, byte(0xBF), content[2])
	assert.True(t, strings.HasPrefix(content[3:], "VLAN,IP Range,Beschreibung,WAN\n"))
}

func TestWriteVlanCSVEmptyDevice(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	require.NoError(t, csvio.WriteVlanCSV(&buf, &model.CommonDevice{}))

	content := strings.TrimPrefix(buf.String(), "\ufeff")
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	assert.Len(t, lines, 1, "empty device: header only")
}

func TestWriteVlanCSVIPRangeDerivation(t *testing.T) {
	t.Parallel()

	dev := &model.CommonDevice{
		VLANs: []model.VLAN{
			{VLANIf: "vlan0.42", Tag: "42", Description: "IT"},
		},
		Interfaces: []model.Interface{
			{Name: "opt1", PhysicalIf: "vlan0.42", IPAddress: "10.42.0.1", Subnet: "24"},
		},
	}

	var buf bytes.Buffer
	require.NoError(t, csvio.WriteVlanCSV(&buf, dev))

	content := strings.TrimPrefix(buf.String(), "\ufeff")
	assert.Contains(t, content, "42,10.42.0.1/24,IT,1")
}
