package opnsense

import (
	"github.com/EvilBit-Labs/opnDossier/pkg/model"
	"github.com/EvilBit-Labs/opnDossier/pkg/schema/opnsense"
)

// SerializeDHCP maps CommonDevice DHCP scopes onto opnsense.Dhcpd keyed by
// interface name.
func SerializeDHCP(scopes []model.DHCPScope) opnsense.Dhcpd {
	items := make(map[string]opnsense.DhcpdInterface, len(scopes))
	for _, s := range scopes {
		out := opnsense.NewDhcpdInterface()
		if s.Enabled {
			out.Enable = "1"
		}
		out.Range = opnsense.Range{From: s.Range.From, To: s.Range.To}
		out.Gateway = s.Gateway
		out.Dnsserver = s.DNSServer
		out.Ntpserver = s.NTPServer
		out.Winsserver = s.WINSServer
		items[s.Interface] = out
	}
	return opnsense.Dhcpd{Items: items}
}
