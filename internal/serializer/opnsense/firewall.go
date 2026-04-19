package opnsense

import (
	"github.com/EvilBit-Labs/opnDossier/pkg/model"
	"github.com/EvilBit-Labs/opnDossier/pkg/schema/opnsense"
)

// SerializeFilter maps CommonDevice firewall rules onto opnsense.Filter.
func SerializeFilter(rules []model.FirewallRule) opnsense.Filter {
	out := opnsense.Filter{}
	for _, r := range rules {
		orule := opnsense.Rule{
			Type:        string(r.Type),
			Descr:       r.Description,
			Interface:   opnsense.InterfaceList(r.Interfaces),
			IPProtocol:  string(r.IPProtocol),
			Direction:   string(r.Direction),
			Protocol:    r.Protocol,
			Source:      endpointToSource(r.Source),
			Destination: endpointToDestination(r.Destination),
			Log:         opnsense.BoolFlag(r.Log),
			Disabled:    opnsense.BoolFlag(r.Disabled),
			Tracker:     r.Tracker,
		}
		out.Rule = append(out.Rule, orule)
	}
	return out
}

// endpointToSource maps model.RuleEndpoint onto opnsense.Source. Only
// Address == "any" becomes the presence-flag form (<any/>). Empty Address
// leaves Source with all match fields unset — OPNsense treats that as
// "no match specified", which is distinct from "explicitly any".
func endpointToSource(ep model.RuleEndpoint) opnsense.Source {
	s := opnsense.Source{Port: ep.Port}
	switch ep.Address {
	case "":
		// Leave Any/Network/Address unset.
	case opnsense.NetworkAny:
		empty := ""
		s.Any = &empty
	default:
		s.Network = ep.Address
	}
	if ep.Negated {
		s.Not = true
	}
	return s
}

// endpointToDestination mirrors endpointToSource for the destination side.
func endpointToDestination(ep model.RuleEndpoint) opnsense.Destination {
	d := opnsense.Destination{Port: ep.Port}
	switch ep.Address {
	case "":
		// Leave Any/Network/Address unset.
	case opnsense.NetworkAny:
		empty := ""
		d.Any = &empty
	default:
		d.Network = ep.Address
	}
	if ep.Negated {
		d.Not = true
	}
	return d
}
