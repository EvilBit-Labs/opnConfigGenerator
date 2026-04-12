package netutil

import (
	"fmt"
	"math/rand/v2"
	"net/netip"
)

// RFC 1918 private address ranges.
var (
	ClassA = netip.MustParsePrefix("10.0.0.0/8")
	ClassB = netip.MustParsePrefix("172.16.0.0/12")
	ClassC = netip.MustParsePrefix("192.168.0.0/16")
)

// IsRFC1918 checks if a prefix is entirely within RFC 1918 space.
func IsRFC1918(prefix netip.Prefix) bool {
	addr := prefix.Addr()
	// Check that the base address falls within one of the RFC 1918 ranges.
	return ClassA.Contains(addr) || ClassB.Contains(addr) || ClassC.Contains(addr)
}

// IsRFC1918Addr checks if a single address is within RFC 1918 space.
func IsRFC1918Addr(addr netip.Addr) bool {
	return ClassA.Contains(addr) || ClassB.Contains(addr) || ClassC.Contains(addr)
}

// GenerateRandomNetwork generates a random /24 RFC 1918 network.
// Weighted distribution: 60% Class A, 25% Class B, 15% Class C.
func GenerateRandomNetwork(rng *rand.Rand) netip.Prefix {
	roll := rng.IntN(100)

	switch {
	case roll < 60:
		return generateClassA(rng)
	case roll < 85:
		return generateClassB(rng)
	default:
		return generateClassC(rng)
	}
}

func generateClassA(rng *rand.Rand) netip.Prefix {
	second := uint8(rng.IntN(254) + 1) // 1-254
	third := uint8(rng.IntN(254) + 1)  // 1-254
	addr := netip.AddrFrom4([4]byte{10, second, third, 0})
	return netip.PrefixFrom(addr, 24)
}

func generateClassB(rng *rand.Rand) netip.Prefix {
	second := uint8(rng.IntN(16) + 16) // 16-31
	third := uint8(rng.IntN(254) + 1)  // 1-254
	addr := netip.AddrFrom4([4]byte{172, second, third, 0})
	return netip.PrefixFrom(addr, 24)
}

func generateClassC(rng *rand.Rand) netip.Prefix {
	third := uint8(rng.IntN(254) + 1) // 1-254
	addr := netip.AddrFrom4([4]byte{192, 168, third, 0})
	return netip.PrefixFrom(addr, 24)
}

// GatewayIP returns the .1 address for a /24 network.
func GatewayIP(network netip.Prefix) netip.Addr {
	base := network.Addr().As4()
	base[3] = 1
	return netip.AddrFrom4(base)
}

// DHCPRangeStart returns the .100 address for a /24 network.
func DHCPRangeStart(network netip.Prefix) netip.Addr {
	base := network.Addr().As4()
	base[3] = 100
	return netip.AddrFrom4(base)
}

// DHCPRangeEnd returns the .200 address for a /24 network.
func DHCPRangeEnd(network netip.Prefix) netip.Addr {
	base := network.Addr().As4()
	base[3] = 200
	return netip.AddrFrom4(base)
}

// HostIP returns a specific host address within a /24 network.
func HostIP(network netip.Prefix, host uint8) netip.Addr {
	base := network.Addr().As4()
	base[3] = host
	return netip.AddrFrom4(base)
}

// ParseRFC1918Network parses and validates a CIDR string as RFC 1918.
func ParseRFC1918Network(s string) (netip.Prefix, error) {
	prefix, err := netip.ParsePrefix(s)
	if err != nil {
		return netip.Prefix{}, fmt.Errorf("parse network %q: %w", s, err)
	}

	if !IsRFC1918(prefix) {
		return netip.Prefix{}, fmt.Errorf("network %s is not RFC 1918 compliant", s)
	}

	return prefix, nil
}
