package faker

import (
	"fmt"
	"math/rand/v2"
	"net/netip"
	"strconv"

	"github.com/EvilBit-Labs/opnConfigGenerator/internal/netutil"
	"github.com/EvilBit-Labs/opnDossier/pkg/model"
	"github.com/brianvoe/gofakeit/v7"
)

const (
	// parentPhysical carries all VLAN-tagged traffic.
	parentPhysical = "igb0"
	// wanPhysical is the upstream/DHCP-facing interface.
	wanPhysical = "igb1"
	// VLAN tag range per 802.1Q: 1 and 4095 are reserved; use [2, 4094].
	vlanTagMin = 2
	vlanTagMax = 4094
	// baseInterfaceCount counts the WAN + LAN pair emitted before any VLANs.
	baseInterfaceCount = 2
)

// networkResult groups everything the network faker produces so callers
// stitch it into a CommonDevice in one step.
type networkResult struct {
	Interfaces []model.Interface
	VLANs      []model.VLAN
}

// fakeNetwork produces the standard WAN + LAN pair plus vlanCount opt
// interfaces, each backed by a unique VLAN tag and RFC 1918 /24 network.
func fakeNetwork(rng *rand.Rand, f *gofakeit.Faker, vlanCount int) networkResult {
	result := networkResult{
		Interfaces: make([]model.Interface, 0, baseInterfaceCount+vlanCount),
		VLANs:      make([]model.VLAN, 0, vlanCount),
	}

	// WAN: DHCP from upstream.
	result.Interfaces = append(result.Interfaces, model.Interface{
		Name:       "wan",
		PhysicalIf: wanPhysical,
		Enabled:    true,
		Type:       "dhcp",
	})

	// LAN: RFC 1918 /24, gateway .1.
	lanNet := netutil.GenerateRandomNetwork(rng)
	lanGW := netutil.GatewayIP(lanNet)
	result.Interfaces = append(result.Interfaces, model.Interface{
		Name:       "lan",
		PhysicalIf: parentPhysical,
		Enabled:    true,
		Type:       "static",
		IPAddress:  lanGW.String(),
		Subnet:     strconv.Itoa(lanNet.Bits()),
	})

	usedTags := make(map[uint16]bool, vlanCount)
	usedNets := map[string]bool{lanNet.String(): true}

	for i := range vlanCount {
		tag := pickUniqueTag(rng, usedTags)
		net := pickUniqueNet(rng, usedNets)
		gw := netutil.GatewayIP(net)

		vlanIf := fmt.Sprintf("vlan0.%d", tag)
		optName := fmt.Sprintf("opt%d", i+1)
		descr := fmt.Sprintf("%s VLAN %d", f.BuzzWord(), tag)

		result.VLANs = append(result.VLANs, model.VLAN{
			VLANIf:      vlanIf,
			PhysicalIf:  parentPhysical,
			Tag:         strconv.FormatUint(uint64(tag), 10),
			Description: descr,
		})
		result.Interfaces = append(result.Interfaces, model.Interface{
			Name:        optName,
			PhysicalIf:  vlanIf,
			Description: descr,
			Enabled:     true,
			Type:        "static",
			IPAddress:   gw.String(),
			Subnet:      strconv.Itoa(net.Bits()),
			Virtual:     true,
		})
	}

	return result
}

// maxPickAttempts bounds the coupon-collector loops below so a shrinking
// pool or an off-by-one in the callers cannot hang the CLI indefinitely.
// Tags have a 4093-slot pool, the RFC 1918 /24 space has ~68K slots — either
// attempts <= 10 * pool-size is far more than a deterministic run needs.
const maxPickAttempts = 100_000

func pickUniqueTag(rng *rand.Rand, used map[uint16]bool) uint16 {
	for range maxPickAttempts {
		//nolint:gosec // Fake data; IntN bounded to uint16 range below.
		tag := uint16(vlanTagMin + rng.IntN(vlanTagMax-vlanTagMin+1))
		if !used[tag] {
			used[tag] = true
			return tag
		}
	}
	panic(fmt.Sprintf("faker: exhausted %d attempts picking a unique VLAN tag (used=%d of %d)",
		maxPickAttempts, len(used), vlanTagMax-vlanTagMin+1))
}

func pickUniqueNet(rng *rand.Rand, used map[string]bool) netip.Prefix {
	for range maxPickAttempts {
		n := netutil.GenerateRandomNetwork(rng)
		if !used[n.String()] {
			used[n.String()] = true
			return n
		}
	}
	panic(fmt.Sprintf("faker: exhausted %d attempts picking a unique RFC 1918 /24 network (used=%d)",
		maxPickAttempts, len(used)))
}
