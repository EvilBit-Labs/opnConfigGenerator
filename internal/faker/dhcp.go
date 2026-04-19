package faker

import (
	"fmt"
	"net/netip"

	"github.com/EvilBit-Labs/opnConfigGenerator/internal/netutil"
	"github.com/EvilBit-Labs/opnDossier/pkg/model"
	"github.com/charmbracelet/log"
)

// fakeDHCPScopes emits one DHCP scope per statically-addressed interface.
// WAN (type="dhcp") is excluded because DHCP servers run on downstream
// interfaces, not on the uplink.
func fakeDHCPScopes(interfaces []model.Interface) []model.DHCPScope {
	scopes := make([]model.DHCPScope, 0, len(interfaces))
	for _, iface := range interfaces {
		if iface.Type != "static" || iface.IPAddress == "" || iface.Subnet == "" {
			continue
		}
		prefix, err := netip.ParsePrefix(fmt.Sprintf("%s/%s", iface.IPAddress, iface.Subnet))
		if err != nil {
			log.Warn("skipping DHCP scope for interface with unparseable prefix",
				"interface", iface.Name, "ip", iface.IPAddress, "subnet", iface.Subnet, "err", err)
			continue
		}
		scopes = append(scopes, model.DHCPScope{
			Interface: iface.Name,
			Enabled:   true,
			Range: model.DHCPRange{
				From: netutil.DHCPRangeStart(prefix).String(),
				To:   netutil.DHCPRangeEnd(prefix).String(),
			},
			Gateway:   iface.IPAddress,
			DNSServer: iface.IPAddress,
		})
	}
	return scopes
}
