package generator

import (
	"fmt"
	"math/rand/v2"
)

// FirewallGenerator produces firewall rules for VLANs at configurable complexity.
type FirewallGenerator struct {
	rng         *rand.Rand
	ruleCounter uint64
	usedTracker map[uint64]bool
}

// NewFirewallGenerator creates a new firewall rule generator.
func NewFirewallGenerator(seed *int64) *FirewallGenerator {
	var rng *rand.Rand
	if seed != nil {
		//nolint:gosec // Deterministic fake data generation, not security-sensitive
		rng = rand.New(rand.NewPCG(uint64(*seed), 0))
	} else {
		//nolint:gosec // Deterministic fake data generation, not security-sensitive
		rng = rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	}

	return &FirewallGenerator{
		rng:         rng,
		usedTracker: make(map[uint64]bool),
	}
}

// GenerateRules produces firewall rules for a VLAN at the given complexity.
func (g *FirewallGenerator) GenerateRules(vlan VlanConfig, complexity FirewallComplexity) []FirewallRule {
	interfaceName := fmt.Sprintf("opt%d", vlan.VlanID)
	vlanNet := vlan.IPNetwork.String()

	var rules []FirewallRule

	rules = append(rules, g.basicRules(vlan.VlanID, interfaceName, vlanNet)...)

	if complexity >= FirewallIntermediate {
		rules = append(rules, g.intermediateRules(vlan.VlanID, interfaceName, vlanNet)...)
	}

	if complexity >= FirewallAdvanced {
		rules = append(rules, g.advancedRules(vlan, interfaceName)...)
	}

	return rules
}

// GenerateRulesForBatch produces rules for multiple VLANs.
func (g *FirewallGenerator) GenerateRulesForBatch(vlans []VlanConfig, complexity FirewallComplexity) []FirewallRule {
	totalRules := len(vlans) * complexity.RulesPerVlan()
	rules := make([]FirewallRule, 0, totalRules)

	for _, vlan := range vlans {
		rules = append(rules, g.GenerateRules(vlan, complexity)...)
	}

	return rules
}

func (g *FirewallGenerator) basicRules(vlanID uint16, iface, vlanNet string) []FirewallRule {
	return []FirewallRule{
		g.newRule(vlanID, iface, vlanNet, vlanNet, "any", "any", "pass", "Allow internal VLAN traffic", false),
		g.newRule(vlanID, iface, vlanNet, "any", "udp", "53", "pass", "Allow DNS queries", false),
		g.newRule(vlanID, iface, vlanNet, "any", "tcp", "80,443", "pass", "Allow HTTP/HTTPS", false),
	}
}

func (g *FirewallGenerator) intermediateRules(vlanID uint16, iface, vlanNet string) []FirewallRule {
	return []FirewallRule{
		g.newRule(vlanID, iface, vlanNet, "any", "udp", "123", "pass", "Allow NTP", false),
		g.newRule(vlanID, iface, vlanNet, "any", "icmp", "any", "pass", "Allow ICMP diagnostics", false),
		g.newRule(vlanID, iface, "any", "any", "tcp", "23,445,3389", "block", "Block common attack ports", false),
		g.newRule(vlanID, iface, "any", "any", "any", "any", "block", "Log denied traffic", true),
	}
}

func (g *FirewallGenerator) advancedRules(vlan VlanConfig, iface string) []FirewallRule {
	vlanNet := vlan.IPNetwork.String()
	rules := make([]FirewallRule, 0, advancedRuleCount-intermediateRuleCount+1)

	switch vlan.Department {
	case DeptIT, DeptEngineering, DeptDevelopment:
		rules = append(rules,
			g.newRule(vlan.VlanID, iface, vlanNet, "any", "tcp", "22", "pass", "Allow SSH access", false),
			g.newRule(vlan.VlanID, iface, vlanNet, "any", "tcp", "3306,5432", "pass", "Allow database access", false),
		)
	case DeptSecurity:
		rules = append(rules,
			g.newRule(vlan.VlanID, iface, vlanNet, "any", "tcp", "514,6514", "pass", "Allow syslog", false),
			g.newRule(vlan.VlanID, iface, vlanNet, "any", "tcp", "9090,9100", "pass", "Allow monitoring", false),
		)
	default:
		rules = append(rules,
			g.newRule(vlan.VlanID, iface, vlanNet, "any", "tcp", "993,587", "pass", "Allow email", false),
			g.newRule(vlan.VlanID, iface, vlanNet, "any", "tcp", "5060,5061", "pass", "Allow SIP/VoIP", false),
		)
	}

	rules = append(rules,
		g.newRule(vlan.VlanID, iface, vlanNet, "any", "tcp", "8080,8443", "pass", "Allow web services", false),
		g.newRule(vlan.VlanID, iface, "any", vlanNet, "tcp", "135,139", "block", "Block SMB/NetBIOS inbound", false),
		g.newRule(vlan.VlanID, iface, vlanNet, "any", "udp", "1194", "pass", "Allow VPN tunnels", false),
		g.newRule(vlan.VlanID, iface, "any", "any", "tcp", "25", "block", "Block outbound SMTP", false),
		g.newRule(vlan.VlanID, iface, vlanNet, "any", "tcp", "1024:65535", "pass", "Allow high ports", false),
		g.newRule(vlan.VlanID, iface, "any", "any", "any", "any", "block", "Default deny all", true),
	)

	return rules
}

func (g *FirewallGenerator) newRule(
	vlanID uint16,
	iface, src, dst, proto, ports, action, desc string,
	log bool,
) FirewallRule {
	g.ruleCounter++
	tracker := g.nextTracker()

	return FirewallRule{
		RuleID:      fmt.Sprintf("rule-%d-%d", vlanID, g.ruleCounter),
		Source:      src,
		Destination: dst,
		Protocol:    proto,
		Ports:       ports,
		Action:      action,
		Direction:   "in",
		Description: desc,
		Log:         log,
		VlanID:      vlanID,
		//nolint:gosec // Capped at max uint16 via min()
		Priority:  uint16(min(g.ruleCounter, uint64(^uint16(0)))),
		Interface: iface,
		Tracker:   tracker,
	}
}

// maxTrackerRetries is the maximum attempts to find a unique tracker value.
const maxTrackerRetries = 1000

func (g *FirewallGenerator) nextTracker() uint64 {
	for range maxTrackerRetries {
		tracker := g.rng.Uint64()
		if !g.usedTracker[tracker] {
			g.usedTracker[tracker] = true
			return tracker
		}
	}

	// Extremely unlikely with uint64 space, but fail deterministically rather than loop forever.
	panic("failed to generate unique tracker after maximum retries — possible RNG issue")
}
