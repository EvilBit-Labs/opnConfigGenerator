// Package opnsensegen handles loading, injecting, and marshaling OPNsense XML configurations.
//
// It uses the opnDossier schema types (github.com/EvilBit-Labs/opnDossier/pkg/schema/opnsense)
// as the canonical OPNsense data model, ensuring consistency across the opnDossier ecosystem.
package opnsensegen

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/EvilBit-Labs/opnConfigGenerator/internal/generator"
	"github.com/EvilBit-Labs/opnConfigGenerator/internal/netutil"
	"github.com/EvilBit-Labs/opnDossier/pkg/schema/opnsense"
)

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

// InjectVlans adds generated VLAN data and corresponding interface entries into the config.
func InjectVlans(cfg *opnsense.OpnSenseDocument, vlans []generator.VlanConfig, optCounter int) {
	if cfg.Interfaces.Items == nil {
		cfg.Interfaces.Items = make(map[string]opnsense.Interface)
	}

	for i, v := range vlans {
		ifName := fmt.Sprintf("opt%d", optCounter+i)
		vlanIfName := fmt.Sprintf("vlan0.%d", v.VlanID)

		cfg.VLANs.VLAN = append(cfg.VLANs.VLAN, opnsense.VLAN{
			If:     "igb0",
			Tag:    strconv.FormatUint(uint64(v.VlanID), 10),
			Descr:  v.Description,
			Vlanif: vlanIfName,
		})

		gateway := netutil.GatewayIP(v.IPNetwork)
		cfg.Interfaces.Items[ifName] = opnsense.Interface{
			Enable: "1",
			Descr:  v.Description,
			If:     vlanIfName,
			IPAddr: gateway.String(),
			Subnet: "24",
		}
	}
}

// InjectDHCP adds generated DHCP configurations into the config.
func InjectDHCP(
	cfg *opnsense.OpnSenseDocument,
	_ []generator.VlanConfig,
	dhcpConfigs []generator.DhcpServerConfig,
	optCounter int,
) {
	if cfg.Dhcpd.Items == nil {
		cfg.Dhcpd.Items = make(map[string]opnsense.DhcpdInterface)
	}

	for i, dhcp := range dhcpConfigs {
		ifName := fmt.Sprintf("opt%d", optCounter+i)

		dhcpIface := opnsense.NewDhcpdInterface()
		dhcpIface.Enable = "1"
		dhcpIface.Range = opnsense.Range{
			From: dhcp.RangeStart.String(),
			To:   dhcp.RangeEnd.String(),
		}
		dhcpIface.Gateway = dhcp.Gateway.String()
		dhcpIface.Dnsserver = strings.Join(dhcp.DNSServers, ",")

		cfg.Dhcpd.Items[ifName] = dhcpIface
	}
}

// InjectFirewallRules adds generated firewall rules into the config.
func InjectFirewallRules(cfg *opnsense.OpnSenseDocument, rules []generator.FirewallRule) {
	for _, r := range rules {
		src := buildSource(r.Source)
		dst := buildDestination(r.Destination, r.Ports)

		var log opnsense.BoolFlag
		if r.Log {
			log = true
		}

		cfg.Filter.Rule = append(cfg.Filter.Rule, opnsense.Rule{
			Type:        r.Action,
			Descr:       r.Description,
			Interface:   opnsense.InterfaceList{r.Interface},
			IPProtocol:  "inet",
			Protocol:    r.Protocol,
			Source:      src,
			Destination: dst,
			Log:         log,
			Direction:   r.Direction,
			Tracker:     strconv.FormatUint(r.Tracker, 10),
		})
	}
}

// MarshalConfig writes the config to XML with proper formatting.
func MarshalConfig(cfg *opnsense.OpnSenseDocument, w io.Writer) error {
	if _, err := io.WriteString(w, xml.Header); err != nil {
		return fmt.Errorf("write XML header: %w", err)
	}

	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")

	if err := enc.Encode(cfg); err != nil {
		return fmt.Errorf("encode config XML: %w", err)
	}

	if _, err := io.WriteString(w, "\n"); err != nil {
		return fmt.Errorf("write trailing newline: %w", err)
	}

	return nil
}

// buildSource creates an opnsense.Source from a generator source string.
func buildSource(source string) opnsense.Source {
	if source == opnsense.NetworkAny {
		empty := ""
		return opnsense.Source{Any: &empty}
	}

	return opnsense.Source{Network: source}
}

// buildDestination creates an opnsense.Destination from generator destination and port strings.
func buildDestination(destination, ports string) opnsense.Destination {
	dst := opnsense.Destination{}

	if destination == opnsense.NetworkAny {
		empty := ""
		dst.Any = &empty
	} else {
		dst.Network = destination
	}

	if ports != opnsense.NetworkAny {
		dst.Port = ports
	}

	return dst
}
