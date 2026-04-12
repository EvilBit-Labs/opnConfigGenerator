package errors_test

import (
	"errors"
	"strings"
	"testing"

	cfgerrors "github.com/EvilBit-Labs/opnConfigGenerator/internal/errors"
)

func TestConfigErrorIs(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		target error
		want   bool
	}{
		{"validation matches", cfgerrors.NewConfigError(cfgerrors.ErrValidation, "bad input"), cfgerrors.ErrValidation, true},
		{"resource exhausted matches", cfgerrors.NewConfigError(cfgerrors.ErrResourceExhausted, "no IDs"), cfgerrors.ErrResourceExhausted, true},
		{"different kind no match", cfgerrors.NewConfigError(cfgerrors.ErrIO, "disk"), cfgerrors.ErrCSV, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := errors.Is(tt.err, tt.target); got != tt.want {
				t.Errorf("errors.Is() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigErrorWrapping(t *testing.T) {
	originalErr := errors.New("original cause")
	configErr := cfgerrors.WrapConfigError(cfgerrors.ErrXML, "processing failed", originalErr)

	// Test that the error unwraps correctly
	if !errors.Is(configErr, cfgerrors.ErrXML) {
		t.Errorf("expected error to match ErrXML")
	}

	// Test error message includes cause
	expected := "XML processing failed: processing failed: original cause"
	if configErr.Error() != expected {
		t.Errorf("Error() = %q, want %q", configErr.Error(), expected)
	}
}

func TestVlanErrorConstructors(t *testing.T) {
	tests := []struct {
		name     string
		err      *cfgerrors.VlanError
		kind     cfgerrors.VlanErrorKind
		contains string
	}{
		{"invalid VLAN ID", cfgerrors.InvalidVlanID(5), cfgerrors.ErrInvalidVlanID, "VLAN ID 5"},
		{"non RFC 1918", cfgerrors.NonRFC1918Network("8.8.8.0/24"), cfgerrors.ErrNonRFC1918Network, "8.8.8.0/24"},
		{"VLAN ID exhausted", cfgerrors.VlanIDExhausted(), cfgerrors.ErrVlanIDExhausted, "exhausted"},
		{"network exhausted", cfgerrors.NetworkExhausted(), cfgerrors.ErrNetworkExhausted, "network pool exhausted"},
		{"invalid WAN", cfgerrors.InvalidWanAssignment(5), cfgerrors.ErrInvalidWanAssignment, "WAN assignment 5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Kind != tt.kind {
				t.Errorf("kind = %v, want %v", tt.err.Kind, tt.kind)
			}
			if !containsStr(tt.err.Error(), tt.contains) {
				t.Errorf("error message %q should contain %q", tt.err.Error(), tt.contains)
			}
		})
	}
}

func TestVlanErrorMessages(t *testing.T) {
	tests := []struct {
		name     string
		createFn func() *cfgerrors.VlanError
		expected string
	}{
		{"VLAN ID range validation", func() *cfgerrors.VlanError { return cfgerrors.InvalidVlanID(9) }, "VLAN ID 9 is outside valid range 10-4094"},
		{"RFC 1918 compliance", func() *cfgerrors.VlanError { return cfgerrors.NonRFC1918Network("192.0.2.0/24") }, "network 192.0.2.0/24 is not RFC 1918 compliant"},
		{"WAN assignment range", func() *cfgerrors.VlanError { return cfgerrors.InvalidWanAssignment(4) }, "WAN assignment 4 is outside valid range 1-3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.createFn()
			if err.Error() != tt.expected {
				t.Errorf("Error() = %q, want %q", err.Error(), tt.expected)
			}
		})
	}
}

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"ErrIO", cfgerrors.ErrIO},
		{"ErrCSV", cfgerrors.ErrCSV},
		{"ErrXML", cfgerrors.ErrXML},
		{"ErrVlanGeneration", cfgerrors.ErrVlanGeneration},
		{"ErrValidation", cfgerrors.ErrValidation},
		{"ErrXMLTemplate", cfgerrors.ErrXMLTemplate},
		{"ErrXMLInjectionNotFound", cfgerrors.ErrXMLInjectionNotFound},
		{"ErrXMLSchemaValidation", cfgerrors.ErrXMLSchemaValidation},
		{"ErrXMLMemoryLimitExceeded", cfgerrors.ErrXMLMemoryLimitExceeded},
		{"ErrConfigNotFound", cfgerrors.ErrConfigNotFound},
		{"ErrInvalidParameter", cfgerrors.ErrInvalidParameter},
		{"ErrResourceExhausted", cfgerrors.ErrResourceExhausted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Error("sentinel error should not be nil")
			}
			if tt.err.Error() == "" {
				t.Error("sentinel error should have non-empty message")
			}
		})
	}
}

func containsStr(s, substr string) bool {
	return strings.Contains(s, substr)
}