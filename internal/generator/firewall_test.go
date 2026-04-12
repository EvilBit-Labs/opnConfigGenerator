package generator_test

import (
	"net/netip"
	"testing"

	"github.com/EvilBit-Labs/opnConfigGenerator/internal/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeVlan(id uint16, dept generator.Department) generator.VlanConfig {
	return generator.VlanConfig{
		VlanID:        id,
		IPNetwork:     netip.MustParsePrefix("10.1.1.0/24"),
		Description:   "test",
		WanAssignment: 1,
		Department:    dept,
	}
}

func TestFirewallRuleCountPerComplexity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		complexity generator.FirewallComplexity
		expected   int
	}{
		{"basic", generator.FirewallBasic, 3},
		{"intermediate", generator.FirewallIntermediate, 7},
		{"advanced", generator.FirewallAdvanced, 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gen := generator.NewFirewallGenerator(new(int64(42)))
			vlan := makeVlan(100, generator.DeptSales)
			rules := gen.GenerateRules(vlan, tt.complexity)
			assert.Len(t, rules, tt.expected)
		})
	}
}

func TestFirewallRuleValidActions(t *testing.T) {
	t.Parallel()

	validActions := map[string]bool{"pass": true, "block": true}

	gen := generator.NewFirewallGenerator(new(int64(42)))
	vlan := makeVlan(100, generator.DeptIT)
	rules := gen.GenerateRules(vlan, generator.FirewallAdvanced)

	for _, rule := range rules {
		assert.True(t, validActions[rule.Action], "rule %s has invalid action: %s", rule.RuleID, rule.Action)
	}
}

func TestFirewallRuleValidProtocols(t *testing.T) {
	t.Parallel()

	validProtocols := map[string]bool{"tcp": true, "udp": true, "icmp": true, "any": true}

	gen := generator.NewFirewallGenerator(new(int64(42)))
	vlan := makeVlan(100, generator.DeptIT)
	rules := gen.GenerateRules(vlan, generator.FirewallAdvanced)

	for _, rule := range rules {
		assert.True(t, validProtocols[rule.Protocol], "rule %s has invalid protocol: %s", rule.RuleID, rule.Protocol)
	}
}

func TestFirewallUniqueTrackers(t *testing.T) {
	t.Parallel()

	gen := generator.NewFirewallGenerator(new(int64(42)))
	vlans := []generator.VlanConfig{
		makeVlan(100, generator.DeptIT),
		makeVlan(200, generator.DeptSales),
		makeVlan(300, generator.DeptEngineering),
	}

	rules := gen.GenerateRulesForBatch(vlans, generator.FirewallAdvanced)

	seen := make(map[uint64]bool)
	for _, rule := range rules {
		assert.False(t, seen[rule.Tracker], "duplicate tracker: %d", rule.Tracker)
		seen[rule.Tracker] = true
	}
}

func TestFirewallUniqueRuleIDs(t *testing.T) {
	t.Parallel()

	gen := generator.NewFirewallGenerator(new(int64(42)))
	vlans := []generator.VlanConfig{
		makeVlan(100, generator.DeptIT),
		makeVlan(200, generator.DeptSales),
	}

	rules := gen.GenerateRulesForBatch(vlans, generator.FirewallIntermediate)

	seen := make(map[string]bool)
	for _, rule := range rules {
		assert.False(t, seen[rule.RuleID], "duplicate rule ID: %s", rule.RuleID)
		seen[rule.RuleID] = true
	}
}

func TestFirewallRulesReferenceCorrectInterface(t *testing.T) {
	t.Parallel()

	gen := generator.NewFirewallGenerator(new(int64(42)))
	vlan := makeVlan(42, generator.DeptIT)
	rules := gen.GenerateRules(vlan, generator.FirewallBasic)

	for _, rule := range rules {
		assert.Equal(t, "opt42", rule.Interface)
	}
}

func TestFirewallBatchTotalRules(t *testing.T) {
	t.Parallel()

	gen := generator.NewFirewallGenerator(new(int64(42)))
	vlans := make([]generator.VlanConfig, 10)
	for i := range vlans {
		vlans[i] = makeVlan(uint16(100+i), generator.DeptSales)
	}

	rules := gen.GenerateRulesForBatch(vlans, generator.FirewallBasic)
	assert.Len(t, rules, 30)
}

func TestFirewallComplexityParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected generator.FirewallComplexity
		wantErr  bool
	}{
		{"basic", generator.FirewallBasic, false},
		{"intermediate", generator.FirewallIntermediate, false},
		{"advanced", generator.FirewallAdvanced, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got, err := generator.ParseFirewallComplexity(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestFirewallAllRulesDirection(t *testing.T) {
	t.Parallel()

	gen := generator.NewFirewallGenerator(new(int64(42)))
	vlan := makeVlan(100, generator.DeptIT)
	rules := gen.GenerateRules(vlan, generator.FirewallAdvanced)

	for _, rule := range rules {
		assert.Equal(t, "in", rule.Direction)
	}
}
