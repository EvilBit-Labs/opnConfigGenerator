// Package opnsensegen is the transport layer for OPNsense XML documents.
// It exposes only load / parse / marshal; all generation and serialization
// logic lives in internal/faker and internal/serializer/opnsense respectively.
//
// It uses the opnDossier schema types (github.com/EvilBit-Labs/opnDossier/pkg/schema/opnsense)
// as the canonical OPNsense data model, ensuring consistency across the opnDossier ecosystem.
package opnsensegen

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/EvilBit-Labs/opnDossier/pkg/schema/opnsense"
)

// mapBackedSections names the OPNsense XML elements whose children are
// serialized from a Go map and therefore emit in non-deterministic order.
// MarshalConfig post-processes these sections to sort children by tag name,
// which is the cheapest way to guarantee byte-stable output without forking
// opnDossier's MarshalXML implementations.
var mapBackedSections = map[string]bool{
	"interfaces": true,
	"dhcpd":      true,
}

// LoadBaseConfig reads and parses a base OPNsense config.xml file.
func LoadBaseConfig(path string) (*opnsense.OpnSenseDocument, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read base config %q: %w", path, err)
	}
	return ParseConfig(data)
}

// ParseConfig parses XML bytes into an OpnSenseDocument.
func ParseConfig(data []byte) (*opnsense.OpnSenseDocument, error) {
	var cfg opnsense.OpnSenseDocument
	if err := xml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config XML: %w", err)
	}
	return &cfg, nil
}

// MarshalConfig writes the config to XML with a standard OPNsense header,
// two-space indentation, and a trailing newline. Children of map-backed
// sections (interfaces, dhcpd) are sorted alphabetically so output is
// byte-stable under a fixed seed.
func MarshalConfig(cfg *opnsense.OpnSenseDocument, w io.Writer) error {
	if _, err := io.WriteString(w, xml.Header); err != nil {
		return fmt.Errorf("write XML header: %w", err)
	}

	var buf bytes.Buffer
	enc := xml.NewEncoder(&buf)
	enc.Indent("", "  ")
	if err := enc.Encode(cfg); err != nil {
		return fmt.Errorf("encode config XML: %w", err)
	}
	if err := enc.Flush(); err != nil {
		return fmt.Errorf("flush XML encoder: %w", err)
	}

	stable, err := sortMapBackedSections(buf.Bytes())
	if err != nil {
		return fmt.Errorf("stabilize XML: %w", err)
	}
	if _, err := w.Write(stable); err != nil {
		return fmt.Errorf("write XML body: %w", err)
	}

	if _, err := io.WriteString(w, "\n"); err != nil {
		return fmt.Errorf("write trailing newline: %w", err)
	}

	return nil
}

// sortMapBackedSections walks the token stream and, whenever it encounters
// a section in mapBackedSections, buffers its direct children, sorts them by
// tag name, and re-emits them in sorted order. Non-section tokens pass
// through untouched.
func sortMapBackedSections(raw []byte) ([]byte, error) {
	dec := xml.NewDecoder(bytes.NewReader(raw))
	var out bytes.Buffer
	enc := xml.NewEncoder(&out)
	enc.Indent("", "  ")

	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}

		start, isStart := tok.(xml.StartElement)
		if !isStart || !mapBackedSections[start.Name.Local] {
			if err := enc.EncodeToken(tok); err != nil {
				return nil, err
			}
			continue
		}

		if err := emitSortedChildren(dec, enc, start); err != nil {
			return nil, err
		}
	}

	if err := enc.Flush(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// emitSortedChildren buffers the direct children of a map-backed section,
// sorts them by start-element name, and emits the section with sorted
// children. It consumes tokens through the section's closing end element.
func emitSortedChildren(dec *xml.Decoder, enc *xml.Encoder, start xml.StartElement) error {
	type child struct {
		name   string
		tokens []xml.Token
	}
	var (
		children []child
		current  child
		depth    int
	)

	for {
		t, err := dec.Token()
		if err != nil {
			return err
		}

		switch tt := t.(type) {
		case xml.StartElement:
			if depth == 0 {
				current = child{name: tt.Name.Local}
			}
			current.tokens = append(current.tokens, xml.CopyToken(t))
			depth++

		case xml.EndElement:
			if depth == 0 {
				// End of the map-backed section itself.
				sort.SliceStable(children, func(i, j int) bool {
					return children[i].name < children[j].name
				})
				if err := enc.EncodeToken(start); err != nil {
					return err
				}
				for _, c := range children {
					for _, ct := range c.tokens {
						if err := enc.EncodeToken(ct); err != nil {
							return err
						}
					}
				}
				return enc.EncodeToken(tt)
			}
			depth--
			current.tokens = append(current.tokens, xml.CopyToken(t))
			if depth == 0 {
				children = append(children, current)
				current = child{}
			}

		default:
			if depth > 0 {
				current.tokens = append(current.tokens, xml.CopyToken(t))
			}
			// CharData/Comment/Directive/ProcInst at depth 0 (between
			// children) is indentation whitespace — drop it; the encoder
			// re-indents on emit.
		}
	}
}
