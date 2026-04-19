package faker

import (
	"fmt"
	"net/netip"

	"github.com/EvilBit-Labs/opnConfigGenerator/internal/netutil"
	"github.com/EvilBit-Labs/opnDossier/pkg/model"
)

// fakeDHCPScopes emits one DHCP scope per statically-addressed interface.
// WAN (type="dhcp") is excluded because DHCP servers run on downstream
// interfaces, not on the uplink. An unparseable IP/subnet pair on any
// interface is a programmer error (the faker or an external CommonDevice
// producer malformed its inputs) and is surfaced as a returned error
// rather than a silent skip.
func fakeDHCPScopes(interfaces []model.Interface) ([]model.DHCPScope, error) {
	scopes := make([]model.DHCPScope, 0, len(interfaces))
	for _, iface := range interfaces {
		if iface.Type != "static" || iface.IPAddress == "" || iface.Subnet == "" {
			continue
		}
		prefix, err := netip.ParsePrefix(fmt.Sprintf("%s/%s", iface.IPAddress, iface.Subnet))
		if err != nil {
			return nil, fmt.Errorf("interface %q has unparseable prefix %s/%s: %w",
				iface.Name, iface.IPAddress, iface.Subnet, err)
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
	return scopes, nil
}
