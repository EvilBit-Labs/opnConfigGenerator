package generator

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	mathrand "math/rand/v2"
	"net/netip"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
)

// VpnType represents the type of VPN configuration.
type VpnType int

// VPN type constants for all supported VPN protocols.
const (
	VpnOpenVPN VpnType = iota
	VpnWireGuard
	VpnIPSec
)

// VpnConfig holds a generated VPN configuration.
type VpnConfig struct {
	ID            string
	Type          VpnType
	Name          string
	TunnelNetwork netip.Prefix
	Port          int
	Protocol      string
	Description   string

	// OpenVPN-specific fields.
	Cipher  string
	TLSAuth bool

	// WireGuard-specific fields.
	PublicKey string
	ListenKey string // redacted in exports

	// IPSec-specific fields.
	IKEVersion int
	DHGroup    int
	HashAlgo   string
}

// VpnGenerator produces VPN configurations with non-overlapping tunnel subnets.
type VpnGenerator struct {
	rng         *mathrand.Rand
	usedSubnets map[netip.Prefix]bool
	tunnelBase  uint8 // Starting third octet for tunnel networks.
}

// NewVpnGenerator creates a new VPN configuration generator.
func NewVpnGenerator(seed *int64) *VpnGenerator {
	var rng *mathrand.Rand
	if seed != nil {
		//nolint:gosec // Deterministic fake data generation, not security-sensitive
		rng = mathrand.New(mathrand.NewPCG(uint64(*seed), 0))
	} else {
		//nolint:gosec // Deterministic fake data generation, not security-sensitive
		rng = mathrand.New(mathrand.NewPCG(mathrand.Uint64(), mathrand.Uint64()))
	}

	return &VpnGenerator{
		rng:         rng,
		usedSubnets: make(map[netip.Prefix]bool),
		tunnelBase:  1,
	}
}

// GenerateConfigs produces VPN configurations.
func (g *VpnGenerator) GenerateConfigs(count int) ([]VpnConfig, error) {
	if count <= 0 {
		return nil, nil
	}

	configs := make([]VpnConfig, 0, count)
	for range count {
		vpnType := VpnType(g.rng.IntN(vpnTypeCount))
		cfg, err := g.generateConfig(vpnType)
		if err != nil {
			return nil, err
		}
		configs = append(configs, cfg)
	}

	return configs, nil
}

// GenerateConfigsOfType produces VPN configurations of a specific type.
func (g *VpnGenerator) GenerateConfigsOfType(vpnType VpnType, count int) ([]VpnConfig, error) {
	if count <= 0 {
		return nil, nil
	}

	configs := make([]VpnConfig, 0, count)
	for range count {
		cfg, err := g.generateConfig(vpnType)
		if err != nil {
			return nil, err
		}
		configs = append(configs, cfg)
	}

	return configs, nil
}

func (g *VpnGenerator) generateConfig(vpnType VpnType) (VpnConfig, error) {
	tunnel, err := g.nextTunnelSubnet()
	if err != nil {
		return VpnConfig{}, err
	}

	switch vpnType {
	case VpnOpenVPN:
		return g.openVPNConfig(tunnel), nil
	case VpnWireGuard:
		return g.wireGuardConfig(tunnel), nil
	case VpnIPSec:
		return g.ipsecConfig(tunnel), nil
	default:
		return g.openVPNConfig(tunnel), nil
	}
}

// VPN generation constants for port ranges and protocol parameters.
const (
	ovpnBasePort    = 1194
	ovpnPortRange   = 100
	wgBasePort      = 51820
	wgPortRange     = 100
	vpnTypeCount    = 3
	ipsecPort       = 500
	ipsecIKEVersion = 2
	ipsecDHGroup    = 14
	maxTunnelOctet  = 254
	tunnelPrefix    = 24
	fakeKeySize     = 32
)

func (g *VpnGenerator) openVPNConfig(tunnel netip.Prefix) VpnConfig {
	port := ovpnBasePort + g.rng.IntN(ovpnPortRange)
	return VpnConfig{
		ID:            uuid.NewString(),
		Type:          VpnOpenVPN,
		Name:          fmt.Sprintf("ovpn-server-%d", port),
		TunnelNetwork: tunnel,
		Port:          port,
		Protocol:      "udp",
		Description:   fmt.Sprintf("OpenVPN server on port %d", port),
		Cipher:        "AES-256-GCM",
		TLSAuth:       true,
	}
}

func (g *VpnGenerator) wireGuardConfig(tunnel netip.Prefix) VpnConfig {
	port := wgBasePort + g.rng.IntN(wgPortRange)
	pubKey := generateFakeKey()
	return VpnConfig{
		ID:            uuid.NewString(),
		Type:          VpnWireGuard,
		Name:          fmt.Sprintf("wg-%d", port),
		TunnelNetwork: tunnel,
		Port:          port,
		Protocol:      "udp",
		Description:   fmt.Sprintf("WireGuard tunnel on port %d", port),
		PublicKey:     pubKey,
		ListenKey:     generateFakeKey(),
	}
}

func (g *VpnGenerator) ipsecConfig(tunnel netip.Prefix) VpnConfig {
	return VpnConfig{
		ID:            uuid.NewString(),
		Type:          VpnIPSec,
		Name:          fmt.Sprintf("ipsec-%s", tunnel.Addr()),
		TunnelNetwork: tunnel,
		Port:          ipsecPort,
		Protocol:      "esp",
		Description:   fmt.Sprintf("IPSec tunnel to %s", tunnel),
		IKEVersion:    ipsecIKEVersion,
		DHGroup:       ipsecDHGroup,
		HashAlgo:      "sha256",
	}
}

// nextTunnelSubnet generates a unique tunnel /24 in the 10.200.x.0 range.
func (g *VpnGenerator) nextTunnelSubnet() (netip.Prefix, error) {
	if g.tunnelBase > maxTunnelOctet {
		return netip.Prefix{}, errors.New("tunnel subnet pool exhausted")
	}

	addr := netip.AddrFrom4([4]byte{10, 200, g.tunnelBase, 0})
	prefix := netip.PrefixFrom(addr, tunnelPrefix)
	g.tunnelBase++
	g.usedSubnets[prefix] = true

	return prefix, nil
}

// UsedSubnets returns all tunnel subnets allocated so far.
func (g *VpnGenerator) UsedSubnets() []netip.Prefix {
	subnets := make([]netip.Prefix, 0, len(g.usedSubnets))
	for s := range g.usedSubnets {
		subnets = append(subnets, s)
	}
	return subnets
}

// generateFakeKey produces a base64-encoded 32-byte fake key.
func generateFakeKey() string {
	key := make([]byte, fakeKeySize)
	if _, err := rand.Read(key); err != nil {
		log.Warn("crypto/rand unavailable, using deterministic placeholder key", "error", err)
		for i := range key {
			key[i] = byte(i)
		}
	}
	return base64.StdEncoding.EncodeToString(key)
}
