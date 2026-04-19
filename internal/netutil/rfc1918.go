// Package netutil provides RFC 1918 network utilities for address generation and validation.
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

// Network generation constants for weighted distribution and addressing.
const (
	distributionRange       = 100 // Total weight range for class distribution
	classAWeight            = 60  // 60% probability for Class A
	classABCumulativeWeight = 85  // 85% cumulative (60% A + 25% B), remainder is C
	subnetPrefix            = 24  // All generated networks use /24 prefix
	maxOctetValue           = 254 // Maximum valid host octet value (1-254)
	classBSecondOctetBase   = 16  // Class B second octet starts at 172.16.x.x
	classBSecondOctetSize   = 16  // Class B range spans 16 values (16-31)
)

// GenerateRandomNetwork generates a random /24 RFC 1918 network.
// Weighted distribution: 60% Class A, 25% Class B, 15% Class C.
func GenerateRandomNetwork(rng *rand.Rand) netip.Prefix {
	roll := rng.IntN(distributionRange)

	switch {
	case roll < classAWeight:
		return generateClassA(rng)
	case roll < classABCumulativeWeight:
		return generateClassB(rng)
	default:
		return generateClassC(rng)
	}
}

func generateClassA(rng *rand.Rand) netip.Prefix {
	//nolint:gosec // IntN(254)+1 yields 1-254, always fits uint8
	second := uint8(rng.IntN(maxOctetValue) + 1)
	//nolint:gosec // IntN(254)+1 yields 1-254, always fits uint8
	third := uint8(rng.IntN(maxOctetValue) + 1)
	addr := netip.AddrFrom4([4]byte{10, second, third, 0})
	return netip.PrefixFrom(addr, subnetPrefix)
}

func generateClassB(rng *rand.Rand) netip.Prefix {
	//nolint:gosec // IntN(16)+16 yields 16-31, always fits uint8
	second := uint8(rng.IntN(classBSecondOctetSize) + classBSecondOctetBase)
	//nolint:gosec // IntN(254)+1 yields 1-254, always fits uint8
	third := uint8(rng.IntN(maxOctetValue) + 1)
	addr := netip.AddrFrom4([4]byte{172, second, third, 0})
	return netip.PrefixFrom(addr, subnetPrefix)
}

func generateClassC(rng *rand.Rand) netip.Prefix {
	//nolint:gosec // IntN(254)+1 yields 1-254, always fits uint8
	third := uint8(rng.IntN(maxOctetValue) + 1)
	addr := netip.AddrFrom4([4]byte{192, 168, third, 0})
	return netip.PrefixFrom(addr, subnetPrefix)
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
