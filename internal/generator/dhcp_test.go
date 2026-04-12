package generator_test

import (
	"math/rand/v2"
	"net/netip"
	"testing"

	"github.com/EvilBit-Labs/opnConfigGenerator/internal/generator"
	"github.com/stretchr/testify/assert"
)

func TestDeriveDHCPConfig(t *testing.T) {
	t.Parallel()

	rng := rand.New(rand.NewPCG(42, 0))
	network := netip.MustParsePrefix("10.42.7.0/24")

	vlan := generator.VlanConfig{
		VlanID:        42,
		IPNetwork:     network,
		Description:   "IT VLAN 42",
		WanAssignment: 1,
		Department:    generator.DeptIT,
	}

	cfg := generator.DeriveDHCPConfig(vlan, rng)

	assert.True(t, cfg.Enabled)
	assert.Equal(t, "10.42.7.1", cfg.Gateway.String())
	assert.Equal(t, "10.42.7.100", cfg.RangeStart.String())
	assert.Equal(t, "10.42.7.200", cfg.RangeEnd.String())
	assert.Equal(t, generator.LeaseTimeCorporate, cfg.LeaseTime)
	assert.Equal(t, "it.local", cfg.DomainName)
	assert.Len(t, cfg.DNSServers, 3)
	assert.Len(t, cfg.NTPServers, 3)
}

func TestDHCPLeaseTimesByDepartment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		dept     generator.Department
		expected int
	}{
		{generator.DeptIT, generator.LeaseTimeCorporate},
		{generator.DeptFinance, generator.LeaseTimeCorporate},
		{generator.DeptLegal, generator.LeaseTimeCorporate},
		{generator.DeptEngineering, generator.LeaseTimeProduction},
		{generator.DeptDevelopment, generator.LeaseTimeProduction},
		{generator.DeptSales, generator.LeaseTimeDynamic},
		{generator.DeptSecurity, generator.LeaseTimeSecurity},
		{generator.DeptHR, generator.LeaseTimeHighMobility},
	}

	for _, tt := range tests {
		t.Run(string(tt.dept), func(t *testing.T) {
			t.Parallel()
			rng := rand.New(rand.NewPCG(42, 0))
			network := netip.MustParsePrefix("10.1.1.0/24")
			vlan := generator.VlanConfig{
				VlanID: 100, IPNetwork: network, Description: "test",
				WanAssignment: 1, Department: tt.dept,
			}
			cfg := generator.DeriveDHCPConfig(vlan, rng)
			assert.Equal(t, tt.expected, cfg.LeaseTime, "department %s", tt.dept)
		})
	}
}

func TestDHCPStaticReservations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		dept          generator.Department
		expectedCount int
	}{
		{generator.DeptIT, 3},
		{generator.DeptEngineering, 2},
		{generator.DeptSecurity, 2},
		{generator.DeptSales, 0},
		{generator.DeptHR, 0},
	}

	for _, tt := range tests {
		t.Run(string(tt.dept), func(t *testing.T) {
			t.Parallel()
			rng := rand.New(rand.NewPCG(42, 0))
			network := netip.MustParsePrefix("10.1.1.0/24")
			vlan := generator.VlanConfig{
				VlanID: 100, IPNetwork: network, Description: "test",
				WanAssignment: 1, Department: tt.dept,
			}
			cfg := generator.DeriveDHCPConfig(vlan, rng)
			assert.Len(t, cfg.StaticReservations, tt.expectedCount)
		})
	}
}

func TestDHCPStaticReservationMACs(t *testing.T) {
	t.Parallel()

	rng := rand.New(rand.NewPCG(42, 0))
	network := netip.MustParsePrefix("10.1.1.0/24")
	vlan := generator.VlanConfig{
		VlanID: 100, IPNetwork: network, Description: "test",
		WanAssignment: 1, Department: generator.DeptIT,
	}
	cfg := generator.DeriveDHCPConfig(vlan, rng)

	for _, res := range cfg.StaticReservations {
		assert.Regexp(t, `^AA:BB:CC:[0-9A-F]{2}:[0-9A-F]{2}:[0-9A-F]{2}$`, res.MAC)
		assert.NotEmpty(t, res.Hostname)
		assert.True(t, network.Contains(res.IP))
	}
}
