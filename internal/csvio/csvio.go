// Package csvio writes VLAN inspection CSVs derived from a *model.CommonDevice.
// The German column headers (VLAN, IP Range, Beschreibung, WAN) are preserved
// from the original tool for compatibility with downstream consumers.
package csvio

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"

	"github.com/EvilBit-Labs/opnDossier/pkg/model"
)

// vlanHeaders are the German column headers this package emits.
var vlanHeaders = []string{"VLAN", "IP Range", "Beschreibung", "WAN"}

// utf8BOM is written first so Excel on Windows detects UTF-8 encoding.
var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// defaultWanAssignment preserves the column shape when the CommonDevice
// model has no concept of WAN assignment. The field is informational only.
const defaultWanAssignment = "1"

// WriteVlanCSV writes the device's VLANs to w in the existing German-header
// CSV format. The column set is: VLAN tag, IP range (derived from the
// matching opt interface's IP and subnet), description, and WAN assignment
// (fixed at "1" — the CommonDevice model has no WAN assignment concept).
func WriteVlanCSV(w io.Writer, device *model.CommonDevice) error {
	if device == nil {
		return errors.New("csvio: device is nil")
	}

	if _, err := w.Write(utf8BOM); err != nil {
		return fmt.Errorf("write BOM: %w", err)
	}

	cw := csv.NewWriter(w)
	if err := cw.Write(vlanHeaders); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	byPhysical := indexByPhysical(device.Interfaces)
	for i, v := range device.VLANs {
		record := []string{v.Tag, ipRangeFor(v, byPhysical), v.Description, defaultWanAssignment}
		if err := cw.Write(record); err != nil {
			return fmt.Errorf("write row %d: %w", i, err)
		}
	}

	cw.Flush()
	return cw.Error()
}

// indexByPhysical returns a map from PhysicalIf to Interface for every
// interface that carries a non-empty PhysicalIf. Interfaces without one are
// skipped because they cannot be matched to a VLAN.VLANIf anyway.
func indexByPhysical(interfaces []model.Interface) map[string]model.Interface {
	byPhysical := make(map[string]model.Interface, len(interfaces))
	for _, iface := range interfaces {
		if iface.PhysicalIf == "" {
			continue
		}
		byPhysical[iface.PhysicalIf] = iface
	}
	return byPhysical
}

// ipRangeFor derives the "<ip>/<subnet>" CSV cell for a VLAN by looking up
// the backing interface in the PhysicalIf index. Returns "" when no match
// exists or the backing interface lacks an IP/Subnet.
func ipRangeFor(v model.VLAN, byPhysical map[string]model.Interface) string {
	iface, ok := byPhysical[v.VLANIf]
	if !ok || iface.IPAddress == "" || iface.Subnet == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s", iface.IPAddress, iface.Subnet)
}
