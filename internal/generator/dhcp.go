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

	//nolint:mnd // Max lease time is conventionally double the base lease time
	maxLease := leaseTime * 2

	return DhcpServerConfig{
		Enabled:            true,
		RangeStart:         rangeStart,
		RangeEnd:           rangeEnd,
		LeaseTime:          leaseTime,
		MaxLeaseTime:       maxLease,
		DNSServers:         dnsServers,
		DomainName:         domainName,
		Gateway:            gateway,
		NTPServers:         defaultNTPServers(),
		StaticReservations: reservations,
	}
}

// staticReservationHostOffset is the starting host IP offset for static DHCP reservations.
const staticReservationHostOffset = 10

func generateStaticReservations(vlan VlanConfig, rng *rand.Rand) []StaticReservation {
	devices := vlan.Department.StaticReservationDevices()
	if len(devices) == 0 {
		return nil
	}

	reservations := make([]StaticReservation, 0, len(devices))
	for i, device := range devices {
		hostIP := netutil.HostIP(vlan.IPNetwork, uint8(staticReservationHostOffset+i))
		mac := generateMAC(rng)
		reservations = append(reservations, StaticReservation{
			MAC:      mac,
			IP:       hostIP,
			Hostname: fmt.Sprintf("%s-%s", domainSafe(string(vlan.Department)), device),
		})
	}

	return reservations
}

// macRandomByteRange is the number of possible values for a random MAC byte (0-255).
const macRandomByteRange = 256

func generateMAC(rng *rand.Rand) string {
	return fmt.Sprintf("AA:BB:CC:%02X:%02X:%02X",
		//nolint:gosec // IntN(256) fits uint8 (0-255), no overflow possible
		uint8(rng.IntN(macRandomByteRange)),
		//nolint:gosec // IntN(256) fits uint8 (0-255), no overflow possible
		uint8(rng.IntN(macRandomByteRange)),
		//nolint:gosec // IntN(256) fits uint8 (0-255), no overflow possible
		uint8(rng.IntN(macRandomByteRange)),
	)
}

func domainSafe(s string) string {
	result := make([]byte, 0, len(s))
	for i := range len(s) {
		c := s[i]
		switch {
		case c >= 'A' && c <= 'Z':
			result = append(result, c+'a'-'A')
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9':
			result = append(result, c)
		case c == ' ':
			result = append(result, '-')
		}
	}
	return string(result)
}
