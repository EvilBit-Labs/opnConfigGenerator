package generator

import (
	"net/netip"
	"time"
)

// VlanConfig holds the generation parameters for a single VLAN.
type VlanConfig struct {
	VlanID        uint16
	IPNetwork     netip.Prefix
	Description   string
	WanAssignment uint8
	Department    Department
}

// DhcpServerConfig holds derived DHCP server settings for a VLAN.
type DhcpServerConfig struct {
	Enabled            bool
	RangeStart         netip.Addr
	RangeEnd           netip.Addr
	LeaseTime          int
	MaxLeaseTime       int
	DNSServers         []string
	DomainName         string
	Gateway            netip.Addr
	NTPServers         []string
	StaticReservations []StaticReservation
}

// StaticReservation maps a MAC address to an IP for DHCP.
type StaticReservation struct {
	MAC      string
	IP       netip.Addr
	Hostname string
}

// FirewallRule holds a single generated firewall rule.
type FirewallRule struct {
	RuleID      string
	Source      string
	Destination string
	Protocol    string
	Ports       string
	Action      string
	Direction   string
	Description string
	Log         bool
	VlanID      uint16
	Priority    uint16
	Interface   string
	Tracker     uint64
}

// FirewallComplexity determines how many rules are generated per VLAN.
type FirewallComplexity int

const (
	// FirewallBasic generates 3 essential rules per VLAN.
	FirewallBasic FirewallComplexity = iota
	// FirewallIntermediate generates 7 rules per VLAN (basic + network services).
	FirewallIntermediate
	// FirewallAdvanced generates 15 rules per VLAN (intermediate + department-specific).
	FirewallAdvanced
)

// Rule count constants for each firewall complexity level.
const (
	basicRuleCount        = 3
	intermediateRuleCount = 7
	advancedRuleCount     = 15
)

// RulesPerVlan returns the number of rules generated for this complexity level.
func (c FirewallComplexity) RulesPerVlan() int {
	switch c {
	case FirewallBasic:
		return basicRuleCount
	case FirewallIntermediate:
		return intermediateRuleCount
	case FirewallAdvanced:
		return advancedRuleCount
	default:
		return basicRuleCount
	}
}

// Firewall complexity level string constants.
const (
	complexityBasic        = "basic"
	complexityIntermediate = "intermediate"
	complexityAdvanced     = "advanced"
)

// String returns the string representation of the complexity level.
func (c FirewallComplexity) String() string {
	switch c {
	case FirewallBasic:
		return complexityBasic
	case FirewallIntermediate:
		return complexityIntermediate
	case FirewallAdvanced:
		return complexityAdvanced
	default:
		return complexityBasic
	}
}

// ParseFirewallComplexity parses a string into a FirewallComplexity.
func ParseFirewallComplexity(s string) (FirewallComplexity, error) {
	switch s {
	case complexityBasic:
		return FirewallBasic, nil
	case complexityIntermediate:
		return FirewallIntermediate, nil
	case complexityAdvanced:
		return FirewallAdvanced, nil
	default:
		return FirewallBasic, &InvalidComplexityError{Value: s}
	}
}

// InvalidComplexityError is returned when a complexity string is not recognized.
type InvalidComplexityError struct {
	Value string
}

func (e *InvalidComplexityError) Error() string {
	return "invalid complexity level '" + e.Value + "': must be basic, intermediate, or advanced"
}

// NatRuleType identifies the kind of NAT rule to generate.
type NatRuleType int

// NatRuleType constants for all supported NAT rule types.
const (
	NatPortForward NatRuleType = iota
	NatSourceNat
	NatDestinationNat
	NatOneToOne
	NatOutbound
)

// WanAssignment controls how VLANs are distributed across WAN interfaces.
type WanAssignment int

const (
	// WanSingle assigns all VLANs to WAN 1.
	WanSingle WanAssignment = iota
	// WanMulti distributes VLANs round-robin across WANs 1-3.
	WanMulti
	// WanBalanced distributes VLANs randomly across WANs 1-3.
	WanBalanced
)

// ParseWanAssignment parses a string into a WanAssignment strategy.
func ParseWanAssignment(s string) (WanAssignment, error) {
	switch s {
	case "single":
		return WanSingle, nil
	case "multi":
		return WanMulti, nil
	case "balanced":
		return WanBalanced, nil
	default:
		return WanSingle, &InvalidWanAssignmentError{Value: s}
	}
}

// InvalidWanAssignmentError is returned when a WAN assignment string is not recognized.
type InvalidWanAssignmentError struct {
	Value string
}

func (e *InvalidWanAssignmentError) Error() string {
	return "invalid WAN assignment '" + e.Value + "': must be single, multi, or balanced"
}

// PerformanceMetrics tracks generation timing and resource usage.
type PerformanceMetrics struct {
	StartTime    time.Time
	Duration     time.Duration
	ConfigCount  int
	MemoryUsedKB int64
}

// NatMapping holds a generated NAT rule.
type NatMapping struct {
	ID          string
	RuleType    NatRuleType
	Interface   string
	Protocol    string
	SourceAddr  string
	SourcePort  string
	DestAddr    string
	DestPort    string
	TargetAddr  string
	TargetPort  string
	Description string
	Log         bool
	VlanID      uint16
}
