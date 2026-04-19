package faker

import (
	"github.com/EvilBit-Labs/opnDossier/pkg/model"
)

// NewCommonDevice returns a fully-populated *model.CommonDevice ready for
// serialization. All randomness is seeded from the Option set; with a fixed
// WithSeed value the function is deterministic.
func NewCommonDevice(opts ...Option) *model.CommonDevice {
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

	net := fakeNetwork(rng, f, cfg.vlanCount)
	dhcp := fakeDHCPScopes(net.Interfaces)

	var fwRules []model.FirewallRule
	if cfg.firewallRules {
		fwRules = fakeFirewallRules(f, net.Interfaces)
	}

	return &model.CommonDevice{
		DeviceType:    model.DeviceTypeOPNsense,
		System:        sys,
		Interfaces:    net.Interfaces,
		VLANs:         net.VLANs,
		DHCP:          dhcp,
		FirewallRules: fwRules,
	}
}
