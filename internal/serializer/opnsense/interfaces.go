package opnsense

import (
	"github.com/EvilBit-Labs/opnDossier/pkg/model"
	"github.com/EvilBit-Labs/opnDossier/pkg/schema/opnsense"
)

// SerializeInterfaces maps a CommonDevice interface slice onto the OPNsense
// map-based Interfaces container, keyed by interface Name (wan, lan, opt*).
func SerializeInterfaces(in []model.Interface) opnsense.Interfaces {
	items := make(map[string]opnsense.Interface, len(in))
	for _, iface := range in {
		out := opnsense.Interface{
			If:    iface.PhysicalIf,
			Descr: iface.Description,
			MTU:   iface.MTU,
		}
		if iface.Enabled {
			out.Enable = "1"
		}
		switch iface.Type {
		case "dhcp":
			out.IPAddr = "dhcp"
		case "static":
			out.IPAddr = iface.IPAddress
			out.Subnet = iface.Subnet
			out.Gateway = iface.Gateway
		}
		items[iface.Name] = out
	}
	return opnsense.Interfaces{Items: items}
}
