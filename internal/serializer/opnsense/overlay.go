package opnsense

import (
	"errors"
	"fmt"

	"github.com/EvilBit-Labs/opnDossier/pkg/model"
	"github.com/EvilBit-Labs/opnDossier/pkg/schema/opnsense"
)

// ErrNilBase is returned when Overlay receives a nil base document.
var ErrNilBase = errors.New("serializer: base document is nil")

// Overlay applies a serialized *model.CommonDevice onto an existing base
// *opnsense.OpnSenseDocument. Fields Phase 1 owns (System, Interfaces,
// VLANs, Dhcpd, Filter) are replaced wholesale. Everything else (Version,
// Theme, Nat, OpenVPN, OPNsense block, Certs, Syslog, ...) is preserved
// from the base. This is the --base-config path: "take this existing config
// and layer my generated content onto it.".
func Overlay(base *opnsense.OpnSenseDocument, device *model.CommonDevice) (*opnsense.OpnSenseDocument, error) {
	if base == nil {
		return nil, ErrNilBase
	}
	serialized, err := Serialize(device)
	if err != nil {
		return nil, fmt.Errorf("overlay: serialize device: %w", err)
	}
	out := *base
	out.System = serialized.System
	out.Interfaces = serialized.Interfaces
	out.VLANs = serialized.VLANs
	out.Dhcpd = serialized.Dhcpd
	out.Filter = serialized.Filter
	return &out, nil
}
