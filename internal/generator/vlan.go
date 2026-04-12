package generator

import (
	"fmt"
	"math/rand/v2"
	"net/netip"

	"github.com/EvilBit-Labs/opnConfigGenerator/internal/netutil"
)

const (
	// MinVlanID is the minimum valid VLAN ID (IEEE 802.1Q).
	MinVlanID = 10
	// MaxVlanID is the maximum valid VLAN ID.
	MaxVlanID = 4094
	// MaxUniqueVlans is the total number of unique VLAN IDs available.
	MaxUniqueVlans = MaxVlanID - MinVlanID + 1
	// maxNetworkRetries is the maximum attempts to find a unique network.
	maxNetworkRetries = 100
)

// VlanGenerator produces unique VLAN configurations with seeded randomness.
type VlanGenerator struct {
	rng         *rand.Rand
	usedVlanIDs map[uint16]bool
	usedNets    map[netip.Prefix]bool
	wanStrategy WanAssignment
	wanCounter  int
}

// NewVlanGenerator creates a new VLAN generator. If seed is nil, a random seed is used.
func NewVlanGenerator(seed *int64, wanStrategy WanAssignment) *VlanGenerator {
	var rng *rand.Rand
	if seed != nil {
		rng = rand.New(rand.NewPCG(uint64(*seed), 0))
	} else {
		rng = rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	}

	return &VlanGenerator{
		rng:         rng,
		usedVlanIDs: make(map[uint16]bool),
		usedNets:    make(map[netip.Prefix]bool),
		wanStrategy: wanStrategy,
	}
}

// GenerateSingle produces a single unique VLAN configuration.
func (g *VlanGenerator) GenerateSingle() (VlanConfig, error) {
	vlanID, err := g.nextUniqueVlanID()
	if err != nil {
		return VlanConfig{}, err
	}

	network, err := g.nextUniqueNetwork()
	if err != nil {
		return VlanConfig{}, err
	}

	dept := RandomDepartment(g.rng)
	wan := g.nextWanAssignment()

	return VlanConfig{
		VlanID:        vlanID,
		IPNetwork:     network,
		Description:   fmt.Sprintf("%s VLAN %d", dept, vlanID),
		WanAssignment: wan,
		Department:    dept,
	}, nil
}

// GenerateBatch produces a batch of unique VLAN configurations.
func (g *VlanGenerator) GenerateBatch(count int) ([]VlanConfig, error) {
	if count <= 0 {
		return nil, fmt.Errorf("count must be positive, got %d", count)
	}

	if count > MaxUniqueVlans {
		return nil, fmt.Errorf(
			"requested %d VLANs exceeds maximum of %d unique VLAN IDs",
			count, MaxUniqueVlans,
		)
	}

	configs := make([]VlanConfig, 0, count)
	for range count {
		cfg, err := g.GenerateSingle()
		if err != nil {
			return nil, fmt.Errorf("generate VLAN %d of %d: %w", len(configs)+1, count, err)
		}
		configs = append(configs, cfg)
	}

	return configs, nil
}

// nextUniqueVlanID generates a unique VLAN ID not already in use.
func (g *VlanGenerator) nextUniqueVlanID() (uint16, error) {
	if len(g.usedVlanIDs) >= MaxUniqueVlans {
		return 0, fmt.Errorf("VLAN ID pool exhausted: all %d IDs in use", MaxUniqueVlans)
	}

	for {
		id := uint16(g.rng.IntN(MaxVlanID-MinVlanID+1)) + MinVlanID
		if !g.usedVlanIDs[id] {
			g.usedVlanIDs[id] = true
			return id, nil
		}
	}
}

// nextUniqueNetwork generates a unique RFC 1918 /24 network.
func (g *VlanGenerator) nextUniqueNetwork() (netip.Prefix, error) {
	for range maxNetworkRetries {
		network := netutil.GenerateRandomNetwork(g.rng)
		canonical := network.Masked()
		if !g.usedNets[canonical] {
			g.usedNets[canonical] = true
			return canonical, nil
		}
	}

	return netip.Prefix{}, fmt.Errorf(
		"failed to generate unique network after %d attempts (%d networks in use)",
		maxNetworkRetries, len(g.usedNets),
	)
}

// nextWanAssignment returns the next WAN assignment based on the strategy.
func (g *VlanGenerator) nextWanAssignment() uint8 {
	switch g.wanStrategy {
	case WanMulti:
		g.wanCounter++
		return uint8((g.wanCounter-1)%3) + 1
	case WanBalanced:
		return uint8(g.rng.IntN(3)) + 1
	default:
		return 1
	}
}

// UsedVlanIDCount returns the number of VLAN IDs already allocated.
func (g *VlanGenerator) UsedVlanIDCount() int {
	return len(g.usedVlanIDs)
}

// UsedNetworkCount returns the number of networks already allocated.
func (g *VlanGenerator) UsedNetworkCount() int {
	return len(g.usedNets)
}
