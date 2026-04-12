// Package errors defines structured error types for the OPNsense Config Generator.
package errors

import (
	"errors"
	"fmt"
)

// Sentinel errors for errors.Is() matching.
var (
	ErrIO                     = errors.New("I/O operation failed")
	ErrCSV                    = errors.New("CSV operation failed")
	ErrXML                    = errors.New("XML processing failed")
	ErrVlanGeneration         = errors.New("VLAN generation failed")
	ErrValidation             = errors.New("validation error")
	ErrXMLTemplate            = errors.New("XML template error")
	ErrXMLInjectionNotFound   = errors.New("XML injection point not found")
	ErrXMLSchemaValidation    = errors.New("XML schema validation failed")
	ErrXMLMemoryLimitExceeded = errors.New("XML memory limit exceeded")
	ErrConfigNotFound         = errors.New("configuration file not found")
	ErrInvalidParameter       = errors.New("invalid parameter")
	ErrResourceExhausted      = errors.New("resource exhausted")
)

// ConfigError provides structured error context for configuration operations.
type ConfigError struct {
	Kind    error
	Message string
	Cause   error
}

func (e *ConfigError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Kind, e.Message, e.Cause)
	}

	return fmt.Sprintf("%s: %s", e.Kind, e.Message)
}

// Unwrap returns the sentinel error kind for errors.Is() matching.
func (e *ConfigError) Unwrap() error {
	return e.Kind
}

// NewConfigError creates a ConfigError with a kind and message.
func NewConfigError(kind error, message string) *ConfigError {
	return &ConfigError{Kind: kind, Message: message}
}

// WrapConfigError creates a ConfigError wrapping an underlying cause.
func WrapConfigError(kind error, message string, cause error) *ConfigError {
	return &ConfigError{Kind: kind, Message: message, Cause: cause}
}

// VlanErrorKind identifies the specific type of VLAN error.
type VlanErrorKind int

const (
	ErrInvalidVlanID VlanErrorKind = iota
	ErrNonRFC1918Network
	ErrNetworkParsing
	ErrVlanIDExhausted
	ErrNetworkExhausted
	ErrInvalidWanAssignment
	ErrInvalidDepartment
	ErrVlanValidationFailed
)

// VlanError represents a VLAN-specific error.
type VlanError struct {
	Kind    VlanErrorKind
	Message string
}

func (e *VlanError) Error() string {
	return e.Message
}

// NewVlanError creates a VlanError with a kind and message.
func NewVlanError(kind VlanErrorKind, message string) *VlanError {
	return &VlanError{Kind: kind, Message: message}
}

// InvalidVlanID creates an error for a VLAN ID outside the valid range.
func InvalidVlanID(id uint16) *VlanError {
	return &VlanError{
		Kind:    ErrInvalidVlanID,
		Message: fmt.Sprintf("VLAN ID %d is outside valid range 10-4094", id),
	}
}

// NonRFC1918Network creates an error for a non-RFC 1918 network.
func NonRFC1918Network(network string) *VlanError {
	return &VlanError{
		Kind:    ErrNonRFC1918Network,
		Message: fmt.Sprintf("network %s is not RFC 1918 compliant", network),
	}
}

// VlanIDExhausted creates an error when all VLAN IDs are used.
func VlanIDExhausted() *VlanError {
	return &VlanError{
		Kind:    ErrVlanIDExhausted,
		Message: "VLAN ID pool exhausted - no more unique IDs available",
	}
}

// NetworkExhausted creates an error when no more unique networks are available.
func NetworkExhausted() *VlanError {
	return &VlanError{
		Kind:    ErrNetworkExhausted,
		Message: "network pool exhausted - no more unique networks available",
	}
}

// InvalidWanAssignment creates an error for a WAN assignment outside 1-3.
func InvalidWanAssignment(wan uint8) *VlanError {
	return &VlanError{
		Kind:    ErrInvalidWanAssignment,
		Message: fmt.Sprintf("WAN assignment %d is outside valid range 1-3", wan),
	}
}
