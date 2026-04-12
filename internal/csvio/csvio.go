// Package csvio handles CSV reading and writing with German headers for VLAN data.
package csvio

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/netip"
	"strconv"

	"github.com/EvilBit-Labs/opnConfigGenerator/internal/generator"
)

// German CSV headers matching the existing Rust implementation.
var vlanHeaders = []string{"VLAN", "IP Range", "Beschreibung", "WAN"}

// UTF-8 BOM for Excel compatibility on Windows.
var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// WriteVlanCSV writes VLAN configurations to a writer in CSV format with German headers.
func WriteVlanCSV(w io.Writer, vlans []generator.VlanConfig) error {
	// Write UTF-8 BOM for Windows/Excel compatibility.
	if _, err := w.Write(utf8BOM); err != nil {
		return fmt.Errorf("write BOM: %w", err)
	}

	cw := csv.NewWriter(w)

	// Write header row.
	if err := cw.Write(vlanHeaders); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	// Write data rows.
	for i, v := range vlans {
		record := []string{
			strconv.FormatUint(uint64(v.VlanID), 10),
			v.IPNetwork.String(),
			v.Description,
			strconv.FormatUint(uint64(v.WanAssignment), 10),
		}
		if err := cw.Write(record); err != nil {
			return fmt.Errorf("write row %d: %w", i, err)
		}
	}

	cw.Flush()
	return cw.Error()
}

// ReadVlanCSV reads VLAN configurations from a CSV reader with German headers.
func ReadVlanCSV(r io.Reader) ([]generator.VlanConfig, error) {
	cr := csv.NewReader(r)

	// Read and validate header.
	header, err := cr.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}

	if err := validateHeader(header); err != nil {
		return nil, err
	}

	var vlans []generator.VlanConfig
	for lineNum := 2; ; lineNum++ {
		record, err := cr.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read line %d: %w", lineNum, err)
		}

		vlan, err := parseVlanRecord(record, lineNum)
		if err != nil {
			return nil, err
		}

		vlans = append(vlans, vlan)
	}

	return vlans, nil
}

func validateHeader(header []string) error {
	if len(header) < len(vlanHeaders) {
		return fmt.Errorf("CSV header has %d columns, expected %d", len(header), len(vlanHeaders))
	}

	// Strip UTF-8 BOM from first field if present (use local copy to avoid mutating caller).
	first := header[0]
	if len(first) >= 3 && first[:3] == string(utf8BOM) {
		first = first[3:]
	}

	for i, expected := range vlanHeaders {
		actual := header[i]
		if i == 0 {
			actual = first
		}
		if actual != expected {
			return fmt.Errorf("CSV header column %d: got %q, expected %q", i, header[i], expected)
		}
	}

	return nil
}

func parseVlanRecord(record []string, lineNum int) (generator.VlanConfig, error) {
	if len(record) < len(vlanHeaders) {
		return generator.VlanConfig{}, fmt.Errorf("line %d: expected 4 columns, got %d", lineNum, len(record))
	}

	vlanID, err := strconv.ParseUint(record[0], 10, 16)
	if err != nil {
		return generator.VlanConfig{}, fmt.Errorf("line %d: invalid VLAN ID %q: %w", lineNum, record[0], err)
	}

	if vlanID < generator.MinVlanID || vlanID > generator.MaxVlanID {
		return generator.VlanConfig{}, fmt.Errorf("line %d: VLAN ID %d outside range %d-%d",
			lineNum, vlanID, generator.MinVlanID, generator.MaxVlanID)
	}

	network, err := netip.ParsePrefix(record[1])
	if err != nil {
		return generator.VlanConfig{}, fmt.Errorf("line %d: invalid network %q: %w", lineNum, record[1], err)
	}

	description := record[2]
	if description == "" {
		return generator.VlanConfig{}, fmt.Errorf("line %d: description cannot be empty", lineNum)
	}

	wan, err := strconv.ParseUint(record[3], 10, 8)
	if err != nil {
		return generator.VlanConfig{}, fmt.Errorf("line %d: invalid WAN assignment %q: %w", lineNum, record[3], err)
	}

	if wan < 1 || wan > 3 {
		return generator.VlanConfig{}, fmt.Errorf("line %d: WAN assignment %d outside range 1-3", lineNum, wan)
	}

	return generator.VlanConfig{
		VlanID:        uint16(vlanID),
		IPNetwork:     network,
		Description:   description,
		WanAssignment: uint8(wan),
	}, nil
}
