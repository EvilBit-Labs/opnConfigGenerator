package faker

import (
	"fmt"

	"github.com/EvilBit-Labs/opnDossier/pkg/model"
)

// MaxVLANCount is the upper bound on WithVLANCount enforced by
// NewCommonDevice. Exported so callers (CLI, tests) can reference the same
// value instead of duplicating the literal.
const MaxVLANCount = 4093

// NewCommonDevice returns a fully-populated *model.CommonDevice ready for
// serialization. All randomness is seeded from the Option set; with a fixed
// WithSeed value the function is deterministic.
//
// Returns an error when the requested VLAN count is out of range or when
// the VLAN tag / RFC 1918 /24 uniqueness pool is exhausted. Callers should
// propagate the error rather than retrying blindly.
func NewCommonDevice(opts ...Option) (*model.CommonDevice, error) {
	cfg := &config{}
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.vlanCount < 0 {
		return nil, fmt.Errorf("VLAN count must be >= 0, got %d", cfg.vlanCount)
	}
	if cfg.vlanCount > MaxVLANCount {
		return nil, fmt.Errorf("VLAN count must be <= %d, got %d", MaxVLANCount, cfg.vlanCount)
	}

	rng, f := newRand(cfg.seed)

	sys := fakeSystem(f)
	if cfg.hostname != "" {
		sys.Hostname = cfg.hostname
	}
	if cfg.domain != "" {
		sys.Domain = cfg.domain
	}

	net, err := fakeNetwork(rng, f, cfg.vlanCount)
	if err != nil {
		return nil, fmt.Errorf("generate network topology: %w", err)
	}

	dhcp, err := fakeDHCPScopes(net.Interfaces)
	if err != nil {
		return nil, fmt.Errorf("generate DHCP scopes: %w", err)
	}

	var fwRules []model.FirewallRule
	if cfg.firewallRules {
		fwRules = fakeFirewallRules(net.Interfaces)
	}

	return &model.CommonDevice{
		DeviceType:    model.DeviceTypeOPNsense,
		System:        sys,
		Interfaces:    net.Interfaces,
		VLANs:         net.VLANs,
		DHCP:          dhcp,
		FirewallRules: fwRules,
	}, nil
}
