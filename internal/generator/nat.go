package generator

import (
	"fmt"
	"math/rand/v2"
	"net/netip"
	"strconv"

	"github.com/EvilBit-Labs/opnConfigGenerator/internal/netutil"
	"github.com/google/uuid"
)

// Common port forward targets.
var portForwardPorts = []struct {
	external string
	internal string
	proto    string
	desc     string
}{
	{"80", "80", "tcp", "HTTP port forward"},
	{"443", "443", "tcp", "HTTPS port forward"},
	{"8080", "8080", "tcp", "Alt HTTP port forward"},
	{"22", "22", "tcp", "SSH port forward"},
	{"3389", "3389", "tcp", "RDP port forward"},
}

// NatGenerator produces NAT rules for VLANs.
type NatGenerator struct {
	rng *rand.Rand
}

// NewNatGenerator creates a new NAT rule generator.
func NewNatGenerator(seed *int64) *NatGenerator {
	var rng *rand.Rand
	if seed != nil {
		rng = rand.New(rand.NewPCG(uint64(*seed), 0))
	} else {
		rng = rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	}

	return &NatGenerator{rng: rng}
}

// GenerateMappings produces NAT rules distributed across the given VLANs.
func (g *NatGenerator) GenerateMappings(vlans []VlanConfig, count int) []NatMapping {
	if count <= 0 || len(vlans) == 0 {
		return nil
	}

	mappings := make([]NatMapping, 0, count)
	for i := range count {
		vlan := vlans[i%len(vlans)]
		ruleType := NatRuleType(g.rng.IntN(5))
		mapping := g.generateMapping(vlan, ruleType)
		mappings = append(mappings, mapping)
	}

	return mappings
}

func (g *NatGenerator) generateMapping(vlan VlanConfig, ruleType NatRuleType) NatMapping {
	gateway := netutil.GatewayIP(vlan.IPNetwork)
	internalHost := netutil.HostIP(vlan.IPNetwork, uint8(g.rng.IntN(200)+10))

	switch ruleType {
	case NatPortForward:
		return g.portForward(vlan, internalHost)
	case NatSourceNat:
		return g.sourceNat(vlan, gateway)
	case NatDestinationNat:
		return g.destinationNat(vlan, internalHost)
	case NatOneToOne:
		return g.oneToOneNat(vlan, internalHost)
	default:
		return g.outboundNat(vlan, gateway)
	}
}

func (g *NatGenerator) portForward(vlan VlanConfig, target netip.Addr) NatMapping {
	pf := portForwardPorts[g.rng.IntN(len(portForwardPorts))]
	return NatMapping{
		ID:          uuid.NewString(),
		RuleType:    NatPortForward,
		Interface:   "wan",
		Protocol:    pf.proto,
		SourceAddr:  "any",
		SourcePort:  "any",
		DestAddr:    "wan_ip",
		DestPort:    pf.external,
		TargetAddr:  target.String(),
		TargetPort:  pf.internal,
		Description: fmt.Sprintf("%s to %s (VLAN %d)", pf.desc, target, vlan.VlanID),
		VlanID:      vlan.VlanID,
	}
}

func (g *NatGenerator) sourceNat(vlan VlanConfig, gateway netip.Addr) NatMapping {
	return NatMapping{
		ID:          uuid.NewString(),
		RuleType:    NatSourceNat,
		Interface:   "wan",
		Protocol:    "any",
		SourceAddr:  vlan.IPNetwork.String(),
		SourcePort:  "any",
		DestAddr:    "any",
		DestPort:    "any",
		TargetAddr:  gateway.String(),
		Description: fmt.Sprintf("Source NAT for %s (VLAN %d)", vlan.IPNetwork, vlan.VlanID),
		VlanID:      vlan.VlanID,
	}
}

func (g *NatGenerator) destinationNat(vlan VlanConfig, target netip.Addr) NatMapping {
	return NatMapping{
		ID:          uuid.NewString(),
		RuleType:    NatDestinationNat,
		Interface:   "wan",
		Protocol:    "tcp",
		SourceAddr:  "any",
		SourcePort:  "any",
		DestAddr:    "wan_ip",
		DestPort:    strconv.Itoa(8000 + g.rng.IntN(1000)),
		TargetAddr:  target.String(),
		TargetPort:  "443",
		Description: fmt.Sprintf("Destination NAT to %s (VLAN %d)", target, vlan.VlanID),
		VlanID:      vlan.VlanID,
	}
}

func (g *NatGenerator) oneToOneNat(vlan VlanConfig, internalIP netip.Addr) NatMapping {
	return NatMapping{
		ID:          uuid.NewString(),
		RuleType:    NatOneToOne,
		Interface:   "wan",
		Protocol:    "any",
		SourceAddr:  internalIP.String(),
		DestAddr:    "any",
		TargetAddr:  "wan_ip",
		Description: fmt.Sprintf("1:1 NAT for %s (VLAN %d)", internalIP, vlan.VlanID),
		VlanID:      vlan.VlanID,
	}
}

func (g *NatGenerator) outboundNat(vlan VlanConfig, gateway netip.Addr) NatMapping {
	return NatMapping{
		ID:          uuid.NewString(),
		RuleType:    NatOutbound,
		Interface:   "wan",
		Protocol:    "any",
		SourceAddr:  vlan.IPNetwork.String(),
		SourcePort:  "any",
		DestAddr:    "any",
		DestPort:    "any",
		TargetAddr:  gateway.String(),
		Description: fmt.Sprintf("Outbound NAT for %s (VLAN %d)", vlan.IPNetwork, vlan.VlanID),
		VlanID:      vlan.VlanID,
	}
}
