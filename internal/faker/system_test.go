package faker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFakeSystemDeterministic(t *testing.T) {
	t.Parallel()

	_, fa := newRand(7)
	_, fb := newRand(7)

	a := fakeSystem(fa)
	b := fakeSystem(fb)

	assert.Equal(t, a, b, "same seed must produce identical System")
}

func TestFakeSystemRequiredFields(t *testing.T) {
	t.Parallel()

	_, f := newRand(7)
	sys := fakeSystem(f)

	assert.NotEmpty(t, sys.Hostname, "Hostname must be populated (validate:hostname)")
	assert.NotEmpty(t, sys.Domain, "Domain must be populated (validate:fqdn)")
	assert.NotEmpty(t, sys.Timezone)
	assert.NotEmpty(t, sys.DNSServers)
	assert.NotEmpty(t, sys.TimeServers)
	assert.Equal(t, "en_US", sys.Language)
}

func TestFakeSystemHostnameIsDNSSafe(t *testing.T) {
	t.Parallel()

	for _, seed := range []int64{1, 2, 3, 42, 100} {
		_, f := newRand(seed)
		sys := fakeSystem(f)

		for _, r := range sys.Hostname {
			ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-'
			assert.Truef(t, ok, "seed %d: hostname char %q not DNS-label-safe in %q", seed, r, sys.Hostname)
		}
	}
}

func TestFakeSystemDomainIsLowercaseFQDN(t *testing.T) {
	t.Parallel()

	for _, seed := range []int64{1, 2, 3, 42, 100} {
		_, f := newRand(seed)
		sys := fakeSystem(f)

		for _, r := range sys.Domain {
			assert.NotContainsf(t, "ABCDEFGHIJKLMNOPQRSTUVWXYZ", string(r),
				"seed %d: domain %q must be lowercase", seed, sys.Domain)
		}
		assert.Containsf(t, sys.Domain, ".",
			"seed %d: domain %q must contain at least one dot", seed, sys.Domain)
	}
}
