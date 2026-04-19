package faker

import (
	"fmt"

	"github.com/EvilBit-Labs/opnDossier/pkg/model"
)

// NewCommonDevice returns a fully-populated *model.CommonDevice ready for
// serialization. All randomness is seeded from the Option set; with a fixed
// WithSeed value the function is deterministic.
//
// Returns an error when the VLAN tag or RFC 1918 /24 uniqueness pool is
// exhausted (e.g., extreme VLAN counts with a small pool). Callers should
// propagate the error to the user rather than retrying blindly.
func NewCommonDevice(opts ...Option) (*model.CommonDevice, error) {
	cfg := &config{}
	for _, opt := range opts {
		opt(cfg)
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
