package generator

import (
	"fmt"
	"math/rand/v2"

	"github.com/EvilBit-Labs/opnConfigGenerator/internal/netutil"
)

// defaultNTPServers returns the NTP servers assigned to DHCP configurations.
func defaultNTPServers() []string {
	return []string{
		"0.pool.ntp.org",
		"1.pool.ntp.org",
		"2.pool.ntp.org",
	}
}

// DeriveDHCPConfig computes a DHCP server configuration from a VLAN config.
func DeriveDHCPConfig(vlan VlanConfig, rng *rand.Rand) DhcpServerConfig {
	gateway := netutil.GatewayIP(vlan.IPNetwork)
	rangeStart := netutil.DHCPRangeStart(vlan.IPNetwork)
	rangeEnd := netutil.DHCPRangeEnd(vlan.IPNetwork)
	leaseTime := vlan.Department.LeaseTime()

	dnsServers := []string{
		gateway.String(),
		"8.8.8.8",
		"1.1.1.1",
	}

	domainName := domainSafe(string(vlan.Department)) + ".local"
	reservations := generateStaticReservations(vlan, rng)

	return DhcpServerConfig{
		Enabled:            true,
		RangeStart:         rangeStart,
		RangeEnd:           rangeEnd,
		LeaseTime:          leaseTime,
		MaxLeaseTime:       leaseTime * 2,
		DNSServers:         dnsServers,
		DomainName:         domainName,
		Gateway:            gateway,
		NTPServers:         defaultNTPServers(),
		StaticReservations: reservations,
	}
}

func generateStaticReservations(vlan VlanConfig, rng *rand.Rand) []StaticReservation {
	devices := vlan.Department.StaticReservationDevices()
	if len(devices) == 0 {
		return nil
	}

	reservations := make([]StaticReservation, 0, len(devices))
	for i, device := range devices {
		hostIP := netutil.HostIP(vlan.IPNetwork, uint8(10+i))
		mac := generateMAC(rng)
		reservations = append(reservations, StaticReservation{
			MAC:      mac,
			IP:       hostIP,
			Hostname: fmt.Sprintf("%s-%s", domainSafe(string(vlan.Department)), device),
		})
	}

	return reservations
}

func generateMAC(rng *rand.Rand) string {
	return fmt.Sprintf("AA:BB:CC:%02X:%02X:%02X",
		uint8(rng.IntN(256)),
		uint8(rng.IntN(256)),
		uint8(rng.IntN(256)),
	)
}

func domainSafe(s string) string {
	result := make([]byte, 0, len(s))
	for i := range len(s) {
		c := s[i]
		switch {
		case c >= 'A' && c <= 'Z':
			result = append(result, c+32)
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9':
			result = append(result, c)
		case c == ' ':
			result = append(result, '-')
		}
	}
	return string(result)
}
