package generator_test

import (
	"net/netip"
	"testing"

	"github.com/EvilBit-Labs/opnConfigGenerator/internal/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNatGenerateMappings(t *testing.T) {
	t.Parallel()

	gen := generator.NewNatGenerator(new(int64(42)))
	vlans := []generator.VlanConfig{
		{VlanID: 100, IPNetwork: netip.MustParsePrefix("10.1.1.0/24"), Department: generator.DeptIT},
		{VlanID: 200, IPNetwork: netip.MustParsePrefix("10.1.2.0/24"), Department: generator.DeptSales},
	}

	mappings := gen.GenerateMappings(vlans, 10)
	assert.Len(t, mappings, 10)
}

func TestNatMappingsDistributeAcrossVlans(t *testing.T) {
	t.Parallel()

	gen := generator.NewNatGenerator(new(int64(42)))
	vlans := []generator.VlanConfig{
		{VlanID: 100, IPNetwork: netip.MustParsePrefix("10.1.1.0/24"), Department: generator.DeptIT},
		{VlanID: 200, IPNetwork: netip.MustParsePrefix("10.1.2.0/24"), Department: generator.DeptSales},
	}

	mappings := gen.GenerateMappings(vlans, 10)
	vlanCounts := make(map[uint16]int)
	for _, m := range mappings {
		vlanCounts[m.VlanID]++
	}

	assert.Equal(t, 5, vlanCounts[100])
	assert.Equal(t, 5, vlanCounts[200])
}

func TestNatMappingsUniqueIDs(t *testing.T) {
	t.Parallel()

	gen := generator.NewNatGenerator(new(int64(42)))
	vlans := []generator.VlanConfig{
		{VlanID: 100, IPNetwork: netip.MustParsePrefix("10.1.1.0/24"), Department: generator.DeptIT},
	}

	mappings := gen.GenerateMappings(vlans, 20)
	seen := make(map[string]bool)
	for _, m := range mappings {
		assert.False(t, seen[m.ID], "duplicate NAT ID: %s", m.ID)
		seen[m.ID] = true
	}
}

func TestNatMappingsHaveDescriptions(t *testing.T) {
	t.Parallel()

	gen := generator.NewNatGenerator(new(int64(42)))
	vlans := []generator.VlanConfig{
		{VlanID: 100, IPNetwork: netip.MustParsePrefix("10.1.1.0/24"), Department: generator.DeptIT},
	}

	mappings := gen.GenerateMappings(vlans, 5)
	for _, m := range mappings {
		assert.NotEmpty(t, m.Description, "mapping %s should have a description", m.ID)
		assert.Contains(t, m.Description, "VLAN 100")
	}
}

func TestNatZeroCountReturnsNil(t *testing.T) {
	t.Parallel()

	gen := generator.NewNatGenerator(new(int64(42)))
	vlans := []generator.VlanConfig{
		{VlanID: 100, IPNetwork: netip.MustParsePrefix("10.1.1.0/24"), Department: generator.DeptIT},
	}

	assert.Nil(t, gen.GenerateMappings(vlans, 0))
	assert.Nil(t, gen.GenerateMappings(vlans, -1))
	assert.Nil(t, gen.GenerateMappings(nil, 5))
}

func TestNatRuleTypes(t *testing.T) {
	t.Parallel()

	gen := generator.NewNatGenerator(new(int64(42)))
	vlans := []generator.VlanConfig{
		{VlanID: 100, IPNetwork: netip.MustParsePrefix("10.1.1.0/24"), Department: generator.DeptIT},
	}

	// Generate enough to get a mix of types.
	mappings := gen.GenerateMappings(vlans, 100)
	require.NotEmpty(t, mappings)

	typeCounts := make(map[generator.NatRuleType]int)
	for _, m := range mappings {
		typeCounts[m.RuleType]++
	}

	// With 100 mappings and 5 types, we should see at least a few of each.
	assert.Greater(t, len(typeCounts), 1, "should generate multiple NAT rule types")
}

func TestNatPortForwardHasValidPorts(t *testing.T) {
	t.Parallel()

	gen := generator.NewNatGenerator(new(int64(42)))
	vlans := []generator.VlanConfig{
		{VlanID: 100, IPNetwork: netip.MustParsePrefix("10.1.1.0/24"), Department: generator.DeptIT},
	}

	mappings := gen.GenerateMappings(vlans, 50)
	for _, m := range mappings {
		if m.RuleType == generator.NatPortForward {
			assert.NotEmpty(t, m.DestPort)
			assert.NotEmpty(t, m.TargetPort)
			assert.NotEmpty(t, m.TargetAddr)
		}
	}
}
