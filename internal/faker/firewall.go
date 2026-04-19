package faker

import (
	"github.com/EvilBit-Labs/opnDossier/pkg/model"
)

// fakeFirewallRules emits a minimal default ruleset: one pass rule per
// non-WAN interface, sourcing from that interface's network to "any". This
// matches OPNsense's out-of-the-box LAN default and is the smallest ruleset
// that produces a semantically useful config.xml.
func fakeFirewallRules(interfaces []model.Interface) []model.FirewallRule {
	rules := make([]model.FirewallRule, 0, len(interfaces))
	for _, iface := range interfaces {
		if iface.Name == "wan" {
			continue
		}
		rules = append(rules, model.FirewallRule{
			Interfaces:  []string{iface.Name},
			Type:        model.RuleTypePass,
			IPProtocol:  model.IPProtocolInet,
			Direction:   model.DirectionIn,
			Description: "Default allow " + iface.Name + " to any",
			Source:      model.RuleEndpoint{Address: iface.Name},
			Destination: model.RuleEndpoint{Address: "any"},
		})
	}
	return rules
}
