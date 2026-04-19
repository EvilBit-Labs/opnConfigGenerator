package faker

import "github.com/EvilBit-Labs/opnDossier/pkg/model"

// Option configures the faker pipeline.
type Option func(*config)

type config struct {
	seed          int64
	vlanCount     int
	firewallRules bool
	hostname      string
	domain        string
	deviceType    model.DeviceType
}

// WithSeed sets a deterministic RNG seed. A seed of 0 is the sentinel for
// "non-deterministic": a fresh random stream per call.
func WithSeed(seed int64) Option {
	return func(c *config) { c.seed = seed }
}

// WithVLANCount requests exactly N VLANs beyond the default WAN/LAN pair.
func WithVLANCount(n int) Option {
	return func(c *config) { c.vlanCount = n }
}

// WithFirewallRules toggles emission of a default firewall ruleset.
func WithFirewallRules(on bool) Option {
	return func(c *config) { c.firewallRules = on }
}

// WithHostname overrides the faker-generated hostname.
func WithHostname(h string) Option {
	return func(c *config) { c.hostname = h }
}

// WithDomain overrides the faker-generated domain.
func WithDomain(d string) Option {
	return func(c *config) { c.domain = d }
}

// WithDeviceType sets the target device type on the produced CommonDevice.
// The faker itself is target-neutral (pure data generation); callers that
// intend to serialize for a specific target set the type here. The default
// is the zero value (unset); callers can pass model.DeviceTypeOPNsense,
// model.DeviceTypePfSense, etc.
func WithDeviceType(dt model.DeviceType) Option {
	return func(c *config) { c.deviceType = dt }
}
