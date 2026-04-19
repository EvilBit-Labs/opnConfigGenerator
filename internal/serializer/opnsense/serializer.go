// Package opnsense serializes a *model.CommonDevice into an
// *opnsense.OpnSenseDocument. This is the inverse of opnDossier's
// pkg/parser/opnsense.ConvertDocument and is the core purpose of
// opnConfigGenerator.
package opnsense

import (
	"errors"

	"github.com/EvilBit-Labs/opnDossier/pkg/model"
	"github.com/EvilBit-Labs/opnDossier/pkg/schema/opnsense"
)

// ErrNilDevice is returned when Serialize is called with a nil input.
var ErrNilDevice = errors.New("serializer: device is nil")

// Serialize converts a *model.CommonDevice into an *opnsense.OpnSenseDocument.
// The returned document is ready for MarshalConfig.
func Serialize(device *model.CommonDevice) (*opnsense.OpnSenseDocument, error) {
	if device == nil {
		return nil, ErrNilDevice
	}
	return &opnsense.OpnSenseDocument{
		Version:    "1.0",
		System:     SerializeSystem(device.System),
		Interfaces: SerializeInterfaces(device.Interfaces),
		VLANs:      SerializeVLANs(device.VLANs),
		Dhcpd:      SerializeDHCP(device.DHCP),
		Filter:     SerializeFilter(device.FirewallRules),
	}, nil
}
