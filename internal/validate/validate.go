// Package validate provides cross-object consistency checks for generated configurations.
package validate

import (
	"fmt"
	"net/netip"

	"github.com/EvilBit-Labs/opnConfigGenerator/internal/generator"
	"github.com/EvilBit-Labs/opnConfigGenerator/internal/netutil"
)

// ValidationResult collects all validation errors for a configuration set.
type ValidationResult struct {
	Errors []string
}

// IsValid returns true if no validation errors were found.
func (r ValidationResult) IsValid() bool {
	return len(r.Errors) == 0
}

// Error returns all validation errors as a single error, or nil if valid.
func (r ValidationResult) Error() error {
	if r.IsValid() {
		return nil
	}
	return fmt.Errorf("validation failed with %d errors: %v", len(r.Errors), r.Errors)
}

// Vlans checks a batch of VLAN configurations for consistency.
func Vlans(vlans []generator.VlanConfig) ValidationResult {
	var result ValidationResult

	seenIDs := make(map[uint16]bool)
	seenNets := make(map[string]bool)

	for i, v := range vlans {
		// VLAN ID range check.
		if v.VlanID < generator.MinVlanID || v.VlanID > generator.MaxVlanID {
			result.Errors = append(
				result.Errors,
				fmt.Sprintf(
					"VLAN[%d]: ID %d outside range %d-%d",
					i,
					v.VlanID,
					generator.MinVlanID,
					generator.MaxVlanID,
				),
			)
		}

		// VLAN ID uniqueness.
		if seenIDs[v.VlanID] {
			result.Errors = append(result.Errors,
				fmt.Sprintf("VLAN[%d]: duplicate VLAN ID %d", i, v.VlanID))
		}
		seenIDs[v.VlanID] = true

		// Network uniqueness.
		netKey := v.IPNetwork.Masked().String()
		if seenNets[netKey] {
			result.Errors = append(result.Errors,
				fmt.Sprintf("VLAN[%d]: duplicate network %s", i, netKey))
		}
		seenNets[netKey] = true

		// RFC 1918 compliance.
		if !netutil.IsRFC1918(v.IPNetwork) {
			result.Errors = append(result.Errors,
				fmt.Sprintf("VLAN[%d]: network %s is not RFC 1918", i, v.IPNetwork))
		}

		// WAN assignment range.
		if v.WanAssignment < 1 || v.WanAssignment > 3 {
			result.Errors = append(result.Errors,
				fmt.Sprintf("VLAN[%d]: WAN assignment %d outside range 1-3", i, v.WanAssignment))
		}

		// Description not empty.
		if v.Description == "" {
			result.Errors = append(result.Errors,
				fmt.Sprintf("VLAN[%d]: description is empty", i))
		}
	}

	return result
}

// FirewallRules checks rules for valid actions, protocols, and interface references.
func FirewallRules(rules []generator.FirewallRule, validInterfaces map[string]bool) ValidationResult {
	var result ValidationResult

	validActions := map[string]bool{"pass": true, "block": true, "reject": true}
	validProtocols := map[string]bool{"tcp": true, "udp": true, "icmp": true, "any": true}
	validDirections := map[string]bool{"in": true, "out": true}

	seenTrackers := make(map[uint64]bool)

	for i, r := range rules {
		if !validActions[r.Action] {
			result.Errors = append(result.Errors,
				fmt.Sprintf("rule[%d] %s: invalid action %q", i, r.RuleID, r.Action))
		}

		if !validProtocols[r.Protocol] {
			result.Errors = append(result.Errors,
				fmt.Sprintf("rule[%d] %s: invalid protocol %q", i, r.RuleID, r.Protocol))
		}

		if !validDirections[r.Direction] {
			result.Errors = append(result.Errors,
				fmt.Sprintf("rule[%d] %s: invalid direction %q", i, r.RuleID, r.Direction))
		}

		if validInterfaces != nil && !validInterfaces[r.Interface] {
			result.Errors = append(result.Errors,
				fmt.Sprintf("rule[%d] %s: references unknown interface %q", i, r.RuleID, r.Interface))
		}

		if seenTrackers[r.Tracker] {
			result.Errors = append(result.Errors,
				fmt.Sprintf("rule[%d] %s: duplicate tracker %d", i, r.RuleID, r.Tracker))
		}
		seenTrackers[r.Tracker] = true
	}

	return result
}

// NoSubnetOverlap checks that VPN tunnel subnets don't overlap with VLAN subnets.
func NoSubnetOverlap(vlanNets, vpnNets []netip.Prefix) ValidationResult {
	var result ValidationResult

	for _, vpn := range vpnNets {
		for _, vlan := range vlanNets {
			if vpn.Overlaps(vlan) {
				result.Errors = append(result.Errors,
					fmt.Sprintf("VPN tunnel %s overlaps with VLAN network %s", vpn, vlan))
			}
		}
	}

	return result
}
