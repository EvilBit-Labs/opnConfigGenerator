package faker

import (
	"testing"

	"github.com/EvilBit-Labs/opnDossier/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFakeFirewallRulesDefaultAllowLAN(t *testing.T) {
	t.Parallel()

	interfaces := []model.Interface{
		{Name: "wan", Type: "dhcp"},
		{Name: "lan", Type: "static"},
	}
	_, f := newRand(1)

	rules := fakeFirewallRules(f, interfaces)

	require.Len(t, rules, 1, "WAN excluded, LAN emits one rule")
	r := rules[0]
	assert.Equal(t, model.RuleTypePass, r.Type)
	assert.Equal(t, []string{"lan"}, r.Interfaces)
	assert.Equal(t, "lan", r.Source.Address)
	assert.Equal(t, "any", r.Destination.Address)
	assert.Equal(t, model.IPProtocolInet, r.IPProtocol)
	assert.Equal(t, model.DirectionIn, r.Direction)
}

func TestFakeFirewallRulesNoInterfacesNoRules(t *testing.T) {
	t.Parallel()

	_, f := newRand(1)
	assert.Empty(t, fakeFirewallRules(f, nil))
}

func TestFakeFirewallRulesOnlyWANNoRules(t *testing.T) {
	t.Parallel()

	_, f := newRand(1)
	rules := fakeFirewallRules(f, []model.Interface{{Name: "wan"}})
	assert.Empty(t, rules)
}
