package netutil_test

import (
	"math/rand/v2"
	"net/netip"
	"testing"

	"github.com/EvilBit-Labs/opnConfigGenerator/internal/netutil"
)

func TestIsRFC1918(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		want   bool
	}{
		// Valid RFC 1918 networks
		{"Class A valid", "10.0.0.0/8", true},
		{"Class A subset", "10.1.0.0/16", true},
		{"Class A /24", "10.5.10.0/24", true},
		{"Class B valid", "172.16.0.0/12", true},
		{"Class B subset", "172.20.0.0/16", true},
		{"Class B /24", "172.25.1.0/24", true},
		{"Class C valid", "192.168.0.0/16", true},
		{"Class C /24", "192.168.1.0/24", true},

		// Invalid networks
		{"Public IP", "8.8.8.0/24", false},
		{"Class A too wide", "9.0.0.0/8", false},
		{"Class B too low", "172.15.0.0/16", false},
		{"Class B too high", "172.32.0.0/16", false},
		{"Class C different", "192.167.0.0/16", false},
		{"Loopback", "127.0.0.0/8", false},
		{"Link-local", "169.254.0.0/16", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix := netip.MustParsePrefix(tt.prefix)
			if got := netutil.IsRFC1918(prefix); got != tt.want {
				t.Errorf("IsRFC1918(%s) = %v, want %v", tt.prefix, got, tt.want)
			}
		})
	}
}

func TestIsRFC1918Addr(t *testing.T) {
	tests := []struct {
		name string
		addr string
		want bool
	}{
		// Valid RFC 1918 addresses
		{"Class A start", "10.0.0.0", true},
		{"Class A mid", "10.100.50.25", true},
		{"Class A end", "10.255.255.255", true},
		{"Class B start", "172.16.0.0", true},
		{"Class B mid", "172.20.10.5", true},
		{"Class B end", "172.31.255.255", true},
		{"Class C start", "192.168.0.0", true},
		{"Class C mid", "192.168.100.50", true},
		{"Class C end", "192.168.255.255", true},

		// Invalid addresses
		{"Public DNS", "8.8.8.8", false},
		{"Class A boundary low", "9.255.255.255", false},
		{"Class A boundary high", "11.0.0.0", false},
		{"Class B boundary low", "172.15.255.255", false},
		{"Class B boundary high", "172.32.0.0", false},
		{"Class C boundary low", "192.167.255.255", false},
		{"Class C boundary high", "192.169.0.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr := netip.MustParseAddr(tt.addr)
			if got := netutil.IsRFC1918Addr(addr); got != tt.want {
				t.Errorf("IsRFC1918Addr(%s) = %v, want %v", tt.addr, got, tt.want)
			}
		})
	}
}

func TestGenerateRandomNetwork(t *testing.T) {
	// Test with seeded RNG for deterministic output
	rng := rand.New(rand.NewPCG(42, 1337))

	// Generate multiple networks and verify they're all RFC 1918
	for range 100 {
		network := netutil.GenerateRandomNetwork(rng)
		if !netutil.IsRFC1918(network) {
			t.Errorf("GenerateRandomNetwork() generated non-RFC 1918 network: %s", network)
		}
		if network.Bits() != 24 {
			t.Errorf("GenerateRandomNetwork() generated non-/24 network: %s", network)
		}
	}
}

func TestGenerateRandomNetworkDistribution(t *testing.T) {
	rng := rand.New(rand.NewPCG(42, 1337))
	const iterations = 1000

	classACount := 0
	classBCount := 0
	classCCount := 0

	for range iterations {
		network := netutil.GenerateRandomNetwork(rng)
		addr := network.Addr()

		if netutil.ClassA.Contains(addr) {
			classACount++
		} else if netutil.ClassB.Contains(addr) {
			classBCount++
		} else if netutil.ClassC.Contains(addr) {
			classCCount++
		}
	}

	// Expected distribution: 60% Class A, 25% Class B, 15% Class C
	// Allow for some variance in random generation
	expectedA := float64(iterations) * 0.60
	expectedB := float64(iterations) * 0.25
	expectedC := float64(iterations) * 0.15

	tolerance := 0.15 // 15% tolerance for random variance

	if abs(float64(classACount)-expectedA) > expectedA*tolerance {
		t.Errorf("Class A distribution: got %d, expected ~%.0f", classACount, expectedA)
	}
	if abs(float64(classBCount)-expectedB) > expectedB*tolerance {
		t.Errorf("Class B distribution: got %d, expected ~%.0f", classBCount, expectedB)
	}
	if abs(float64(classCCount)-expectedC) > expectedC*tolerance {
		t.Errorf("Class C distribution: got %d, expected ~%.0f", classCCount, expectedC)
	}
}

func TestGatewayIP(t *testing.T) {
	tests := []struct {
		name     string
		network  string
		expected string
	}{
		{"Class A", "10.5.10.0/24", "10.5.10.1"},
		{"Class B", "172.20.30.0/24", "172.20.30.1"},
		{"Class C", "192.168.1.0/24", "192.168.1.1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			network := netip.MustParsePrefix(tt.network)
			expected := netip.MustParseAddr(tt.expected)
			got := netutil.GatewayIP(network)
			if got != expected {
				t.Errorf("GatewayIP(%s) = %s, want %s", tt.network, got, expected)
			}
		})
	}
}

func TestDHCPRangeStart(t *testing.T) {
	tests := []struct {
		name     string
		network  string
		expected string
	}{
		{"Class A", "10.5.10.0/24", "10.5.10.100"},
		{"Class B", "172.20.30.0/24", "172.20.30.100"},
		{"Class C", "192.168.1.0/24", "192.168.1.100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			network := netip.MustParsePrefix(tt.network)
			expected := netip.MustParseAddr(tt.expected)
			got := netutil.DHCPRangeStart(network)
			if got != expected {
				t.Errorf("DHCPRangeStart(%s) = %s, want %s", tt.network, got, expected)
			}
		})
	}
}

func TestDHCPRangeEnd(t *testing.T) {
	tests := []struct {
		name     string
		network  string
		expected string
	}{
		{"Class A", "10.5.10.0/24", "10.5.10.200"},
		{"Class B", "172.20.30.0/24", "172.20.30.200"},
		{"Class C", "192.168.1.0/24", "192.168.1.200"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			network := netip.MustParsePrefix(tt.network)
			expected := netip.MustParseAddr(tt.expected)
			got := netutil.DHCPRangeEnd(network)
			if got != expected {
				t.Errorf("DHCPRangeEnd(%s) = %s, want %s", tt.network, got, expected)
			}
		})
	}
}

func TestHostIP(t *testing.T) {
	tests := []struct {
		name     string
		network  string
		host     uint8
		expected string
	}{
		{"Host 50", "10.5.10.0/24", 50, "10.5.10.50"},
		{"Host 1", "172.20.30.0/24", 1, "172.20.30.1"},
		{"Host 254", "192.168.1.0/24", 254, "192.168.1.254"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			network := netip.MustParsePrefix(tt.network)
			expected := netip.MustParseAddr(tt.expected)
			got := netutil.HostIP(network, tt.host)
			if got != expected {
				t.Errorf("HostIP(%s, %d) = %s, want %s", tt.network, tt.host, got, expected)
			}
		})
	}
}

func TestParseRFC1918Network(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		shouldErr bool
		expected  string
	}{
		{"valid Class A", "10.5.10.0/24", false, "10.5.10.0/24"},
		{"valid Class B", "172.20.30.0/24", false, "172.20.30.0/24"},
		{"valid Class C", "192.168.1.0/24", false, "192.168.1.0/24"},
		{"invalid format", "not.a.network", true, ""},
		{"public network", "8.8.8.0/24", true, ""},
		{"non-RFC1918", "203.0.113.0/24", true, ""},
		{"invalid Class B", "172.32.1.0/24", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := netutil.ParseRFC1918Network(tt.input)
			if tt.shouldErr {
				if err == nil {
					t.Errorf("ParseRFC1918Network(%s) expected error but got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("ParseRFC1918Network(%s) unexpected error: %v", tt.input, err)
				}
				expected := netip.MustParsePrefix(tt.expected)
				if got != expected {
					t.Errorf("ParseRFC1918Network(%s) = %s, want %s", tt.input, got, expected)
				}
			}
		})
	}
}

func TestDeterministicGeneration(t *testing.T) {
	// Test that the same seed produces the same sequence
	seed1 := uint64(12345)
	seed2 := uint64(67890)

	rng1a := rand.New(rand.NewPCG(seed1, seed2))
	rng1b := rand.New(rand.NewPCG(seed1, seed2))

	network1a := netutil.GenerateRandomNetwork(rng1a)
	network1b := netutil.GenerateRandomNetwork(rng1b)

	if network1a != network1b {
		t.Errorf("Same seed should produce same network: got %s and %s", network1a, network1b)
	}

	// Test that different seeds produce different sequences
	// Generate multiple values to increase chance of difference
	same := true
	for i := range 10 {
		rng1c := rand.New(rand.NewPCG(seed1, seed2))
		rng2c := rand.New(rand.NewPCG(seed2, seed1))

		// Skip to position i
		for range i {
			netutil.GenerateRandomNetwork(rng1c)
			netutil.GenerateRandomNetwork(rng2c)
		}

		net1 := netutil.GenerateRandomNetwork(rng1c)
		net2 := netutil.GenerateRandomNetwork(rng2c)

		if net1 != net2 {
			same = false
			break
		}
	}

	if same {
		t.Error("Different seeds should eventually produce different networks")
	}
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}
