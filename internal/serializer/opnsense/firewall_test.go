package opnsense_test

import (
	"testing"

	serializer "github.com/EvilBit-Labs/opnConfigGenerator/internal/serializer/opnsense"
	"github.com/EvilBit-Labs/opnDossier/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSerializeFilterPassRule(t *testing.T) {
	t.Parallel()

	in := []model.FirewallRule{{
		Interfaces:  []string{"lan"},
		Type:        model.RuleTypePass,
		Description: "Default allow LAN to any",
		IPProtocol:  model.IPProtocolInet,
		Direction:   model.DirectionIn,
		Source:      model.RuleEndpoint{Address: "lan"},
		Destination: model.RuleEndpoint{Address: "any"},
	}}

	out := serializer.SerializeFilter(in)

	require.Len(t, out.Rule, 1)
	r := out.Rule[0]
	assert.Equal(t, "pass", r.Type)
	assert.Equal(t, "Default allow LAN to any", r.Descr)
	assert.Equal(t, "inet", r.IPProtocol)
	assert.Equal(t, "in", r.Direction)
	assert.Equal(t, "lan", r.Source.Network)
	require.NotNil(t, r.Destination.Any, "destination 'any' must produce a non-nil Any pointer")
}

func TestSerializeFilterEmpty(t *testing.T) {
	t.Parallel()

	assert.Empty(t, serializer.SerializeFilter(nil).Rule)
}

func TestSerializeFilterEmptyEndpointStaysEmpty(t *testing.T) {
	t.Parallel()

	in := []model.FirewallRule{{
		Interfaces:  []string{"lan"},
		Type:        model.RuleTypePass,
		Source:      model.RuleEndpoint{},
		Destination: model.RuleEndpoint{Address: "10.0.0.0/24"},
	}}

	out := serializer.SerializeFilter(in)
	require.Len(t, out.Rule, 1)
	// Empty source (Address=="") is distinct from "any"; Any pointer stays nil.
	assert.Nil(t, out.Rule[0].Source.Any)
	assert.Empty(t, out.Rule[0].Source.Network)
	assert.Empty(t, out.Rule[0].Source.Address)
	assert.Equal(t, "10.0.0.0/24", out.Rule[0].Destination.Network)
}

func TestSerializeFilterExplicitAnyEmitsAnyElement(t *testing.T) {
	t.Parallel()

	in := []model.FirewallRule{{
		Interfaces:  []string{"lan"},
		Type:        model.RuleTypePass,
		Source:      model.RuleEndpoint{Address: "lan"},
		Destination: model.RuleEndpoint{Address: "any"},
	}}

	out := serializer.SerializeFilter(in)
	require.Len(t, out.Rule, 1)
	assert.Equal(t, "lan", out.Rule[0].Source.Network)
	assert.Nil(t, out.Rule[0].Source.Any, "explicit source network must not emit <any/>")
	require.NotNil(t, out.Rule[0].Destination.Any, "explicit 'any' destination must emit <any/>")
}
