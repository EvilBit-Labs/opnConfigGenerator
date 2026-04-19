package validate_test

import (
	"net/netip"
	"testing"

	"github.com/EvilBit-Labs/opnConfigGenerator/internal/generator"
	"github.com/EvilBit-Labs/opnConfigGenerator/internal/validate"
	"github.com/stretchr/testify/assert"
)

func TestValidateVlansValid(t *testing.T) {
	t.Parallel()

	vlans := []generator.VlanConfig{
		{VlanID: 100, IPNetwork: netip.MustParsePrefix("10.1.1.0/24"), Description: "IT VLAN", WanAssignment: 1},
		{VlanID: 200, IPNetwork: netip.MustParsePrefix("10.1.2.0/24"), Description: "Sales VLAN", WanAssignment: 2},
	}

	result := validate.Vlans(vlans)
	assert.True(t, result.IsValid(), "valid VLANs should pass: %v", result.Errors)
}

func TestValidateVlansDuplicateIDs(t *testing.T) {
	t.Parallel()

	vlans := []generator.VlanConfig{
		{VlanID: 100, IPNetwork: netip.MustParsePrefix("10.1.1.0/24"), Description: "test", WanAssignment: 1},
		{VlanID: 100, IPNetwork: netip.MustParsePrefix("10.1.2.0/24"), Description: "test", WanAssignment: 1},
	}

	result := validate.Vlans(vlans)
	assert.False(t, result.IsValid())
	assert.Contains(t, result.Errors[0], "duplicate VLAN ID")
}

func TestValidateVlansDuplicateNetworks(t *testing.T) {
	t.Parallel()

	vlans := []generator.VlanConfig{
		{VlanID: 100, IPNetwork: netip.MustParsePrefix("10.1.1.0/24"), Description: "test", WanAssignment: 1},
		{VlanID: 200, IPNetwork: netip.MustParsePrefix("10.1.1.0/24"), Description: "test", WanAssignment: 1},
	}

	result := validate.Vlans(vlans)
	assert.False(t, result.IsValid())
	assert.Contains(t, result.Errors[0], "duplicate network")
}

func TestValidateVlansIDOutOfRange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		vlanID  uint16
		wantErr bool
	}{
		{"below minimum", 5, true},
		{"at minimum", 10, false},
		{"at maximum", 4094, false},
		{"above maximum", 4095, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			vlans := []generator.VlanConfig{
				{
					VlanID:        tt.vlanID,
					IPNetwork:     netip.MustParsePrefix("10.1.1.0/24"),
					Description:   "test",
					WanAssignment: 1,
				},
			}
			result := validate.Vlans(vlans)
			if tt.wantErr {
				assert.False(t, result.IsValid())
			} else {
				assert.True(t, result.IsValid(), "errors: %v", result.Errors)
			}
		})
	}
}

func TestValidateVlansNonRFC1918(t *testing.T) {
	t.Parallel()

	vlans := []generator.VlanConfig{
		{VlanID: 100, IPNetwork: netip.MustParsePrefix("8.8.8.0/24"), Description: "test", WanAssignment: 1},
	}

	result := validate.Vlans(vlans)
	assert.False(t, result.IsValid())
	assert.Contains(t, result.Errors[0], "not RFC 1918")
}

func TestValidateVlansInvalidWan(t *testing.T) {
	t.Parallel()

	vlans := []generator.VlanConfig{
		{VlanID: 100, IPNetwork: netip.MustParsePrefix("10.1.1.0/24"), Description: "test", WanAssignment: 0},
	}

	result := validate.Vlans(vlans)
	assert.False(t, result.IsValid())
	assert.Contains(t, result.Errors[0], "WAN assignment")
}

func TestValidateVlansEmptyDescription(t *testing.T) {
	t.Parallel()

	vlans := []generator.VlanConfig{
		{VlanID: 100, IPNetwork: netip.MustParsePrefix("10.1.1.0/24"), Description: "", WanAssignment: 1},
	}

	result := validate.Vlans(vlans)
	assert.False(t, result.IsValid())
	assert.Contains(t, result.Errors[0], "description is empty")
}

func TestValidateFirewallRulesValid(t *testing.T) {
	t.Parallel()

	rules := []generator.FirewallRule{
		{RuleID: "r1", Action: "pass", Protocol: "tcp", Direction: "in", Interface: "opt1", Tracker: 1},
		{RuleID: "r2", Action: "block", Protocol: "udp", Direction: "in", Interface: "opt1", Tracker: 2},
	}

	interfaces := map[string]bool{"opt1": true}
	result := validate.FirewallRules(rules, interfaces)
	assert.True(t, result.IsValid(), "errors: %v", result.Errors)
}

func TestValidateFirewallRulesInvalidAction(t *testing.T) {
	t.Parallel()

	rules := []generator.FirewallRule{
		{RuleID: "r1", Action: "deny", Protocol: "tcp", Direction: "in", Interface: "opt1", Tracker: 1},
	}

	result := validate.FirewallRules(rules, nil)
	assert.False(t, result.IsValid())
	assert.Contains(t, result.Errors[0], "invalid action")
}

func TestValidateFirewallRulesDuplicateTracker(t *testing.T) {
	t.Parallel()

	rules := []generator.FirewallRule{
		{RuleID: "r1", Action: "pass", Protocol: "tcp", Direction: "in", Tracker: 42},
		{RuleID: "r2", Action: "pass", Protocol: "tcp", Direction: "in", Tracker: 42},
	}

	result := validate.FirewallRules(rules, nil)
	assert.False(t, result.IsValid())
	assert.Contains(t, result.Errors[0], "duplicate tracker")
}

func TestValidateFirewallRulesUnknownInterface(t *testing.T) {
	t.Parallel()

	rules := []generator.FirewallRule{
		{RuleID: "r1", Action: "pass", Protocol: "tcp", Direction: "in", Interface: "opt99", Tracker: 1},
	}

	interfaces := map[string]bool{"opt1": true, "opt2": true}
	result := validate.FirewallRules(rules, interfaces)
	assert.False(t, result.IsValid())
	assert.Contains(t, result.Errors[0], "unknown interface")
}

func TestValidateNoSubnetOverlap(t *testing.T) {
	t.Parallel()

	vlanNets := []netip.Prefix{
		netip.MustParsePrefix("10.1.1.0/24"),
		netip.MustParsePrefix("10.1.2.0/24"),
	}

	// Non-overlapping VPN subnets.
	vpnNets := []netip.Prefix{
		netip.MustParsePrefix("10.200.1.0/24"),
		netip.MustParsePrefix("10.200.2.0/24"),
	}

	result := validate.NoSubnetOverlap(vlanNets, vpnNets)
	assert.True(t, result.IsValid(), "non-overlapping should be valid: %v", result.Errors)
}

func TestValidateSubnetOverlapDetected(t *testing.T) {
	t.Parallel()

	vlanNets := []netip.Prefix{
		netip.MustParsePrefix("10.1.1.0/24"),
	}

	vpnNets := []netip.Prefix{
		netip.MustParsePrefix("10.1.1.0/24"), // Same as VLAN!
	}

	result := validate.NoSubnetOverlap(vlanNets, vpnNets)
	assert.False(t, result.IsValid())
	assert.Contains(t, result.Errors[0], "overlaps")
}
