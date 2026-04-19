package opnsense

import (
	"github.com/EvilBit-Labs/opnDossier/pkg/model"
	"github.com/EvilBit-Labs/opnDossier/pkg/schema/opnsense"
)

// virtualIfaceMarker is the opnsense.Interface.Virtual sentinel meaning
// "virtual interface" (int-valued in the opnDossier schema; 1 == virtual).
const virtualIfaceMarker = 1

// SerializeInterfaces maps a CommonDevice interface slice onto the OPNsense
// map-based Interfaces container, keyed by interface Name (wan, lan, opt*).
//
// Type and Virtual are propagated verbatim because opnDossier's ConvertDocument
// reads them back on the reverse trip (iface.Virtual != 0, iface.Type
// verbatim). Dropping either silently breaks round-trip parity.
func SerializeInterfaces(in []model.Interface) opnsense.Interfaces {
	items := make(map[string]opnsense.Interface, len(in))
	for _, iface := range in {
		out := opnsense.Interface{
			If:    iface.PhysicalIf,
			Descr: iface.Description,
			MTU:   iface.MTU,
			Type:  iface.Type,
		}
		if iface.Enabled {
			out.Enable = "1"
		}
		if iface.Virtual {
			out.Virtual = virtualIfaceMarker
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
