// Package xmlgen handles loading, injecting, and marshaling OPNsense XML configurations.
package xmlgen

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/EvilBit-Labs/opnConfigGenerator/internal/generator"
	"github.com/EvilBit-Labs/opnConfigGenerator/internal/netutil"
)

// OpnSenseConfig is a simplified representation of the config.xml structure
// for injection purposes. In production, this would use opnDossier's schema types.
type OpnSenseConfig struct {
	XMLName    xml.Name        `xml:"opnsense"`
	Version    string          `xml:"version"`
	System     SystemConfig    `xml:"system"`
	VLANs      VLANSection     `xml:"vlans"`
	Interfaces InterfaceSection `xml:"interfaces"`
	Dhcpd      DhcpdSection    `xml:"dhcpd"`
	Filter     FilterSection   `xml:"filter"`
	Nat        NatSection      `xml:"nat"`
	OpenVPN    OpenVPNSection  `xml:"openvpn"`
	WireGuard  WireGuardSection `xml:"wireguard"`
	IPSec      IPSecSection    `xml:"ipsec"`
	Extra      []XMLExtra      `xml:",any"`
}

// SystemConfig holds basic system configuration.
type SystemConfig struct {
	Hostname string `xml:"hostname"`
	Domain   string `xml:"domain"`
	Extra    []XMLExtra `xml:",any"`
}

// XMLExtra captures any unrecognized XML elements for round-trip preservation.
type XMLExtra struct {
	XMLName xml.Name
	Content []byte `xml:",innerxml"`
}

// VLANSection holds VLAN definitions.
type VLANSection struct {
	VLANs []VLANEntry `xml:"vlan"`
}

// VLANEntry is a single VLAN definition in the config.
type VLANEntry struct {
	If     string `xml:"if"`
	Tag    uint16 `xml:"tag"`
	Descr  string `xml:"descr"`
	VlanIf string `xml:"vlanif"`
}

// InterfaceSection holds interface definitions.
type InterfaceSection struct {
	Entries []InterfaceEntry `xml:",any"`
}

// InterfaceEntry is a single interface.
type InterfaceEntry struct {
	XMLName xml.Name
	Enable  string `xml:"enable,omitempty"`
	Descr   string `xml:"descr,omitempty"`
	If      string `xml:"if,omitempty"`
	IPAddr  string `xml:"ipaddr,omitempty"`
	Subnet  string `xml:"subnet,omitempty"`
}

// DhcpdSection holds DHCP configurations.
type DhcpdSection struct {
	Entries []DhcpdEntry `xml:",any"`
}

// DhcpdEntry is DHCP config for a single interface.
type DhcpdEntry struct {
	XMLName xml.Name
	Enable  string         `xml:"enable,omitempty"`
	Range   DHCPRange      `xml:"range,omitempty"`
	Gateway string         `xml:"gateway,omitempty"`
	Domain  string         `xml:"domain,omitempty"`
	DNS1    string         `xml:"dnsserver,omitempty"`
	Lease   string         `xml:"defaultleasetime,omitempty"`
	MaxLease string        `xml:"maxleasetime,omitempty"`
}

// DHCPRange defines a DHCP address range.
type DHCPRange struct {
	From string `xml:"from"`
	To   string `xml:"to"`
}

// FilterSection holds firewall filter rules.
type FilterSection struct {
	Rules []FilterRule `xml:"rule"`
}

// FilterRule is a single firewall rule.
type FilterRule struct {
	Type        string `xml:"type"`
	Descr       string `xml:"descr"`
	Interface   string `xml:"interface"`
	IPProtocol  string `xml:"ipprotocol"`
	Protocol    string `xml:"protocol,omitempty"`
	Source      RuleSrc `xml:"source"`
	Destination RuleDst `xml:"destination"`
	Log         string `xml:"log,omitempty"`
	Direction   string `xml:"direction,omitempty"`
	Tracker     string `xml:"tracker"`
}

// RuleSrc is the source specification of a firewall rule.
type RuleSrc struct {
	Network string `xml:"network,omitempty"`
	Any     string `xml:"any,omitempty"`
}

// RuleDst is the destination specification of a firewall rule.
type RuleDst struct {
	Network string `xml:"network,omitempty"`
	Any     string `xml:"any,omitempty"`
	Port    string `xml:"port,omitempty"`
}

// NatSection holds NAT rules.
type NatSection struct {
	Outbound OutboundNat `xml:"outbound,omitempty"`
	Rules    []NatRule   `xml:"rule,omitempty"`
}

// OutboundNat holds outbound NAT configuration.
type OutboundNat struct {
	Mode  string    `xml:"mode,omitempty"`
	Rules []NatRule `xml:"rule,omitempty"`
}

// NatRule is a single NAT rule.
type NatRule struct {
	Descr     string `xml:"descr"`
	Interface string `xml:"interface"`
	Protocol  string `xml:"protocol,omitempty"`
	Source    string `xml:"source,omitempty"`
	Target   string `xml:"target,omitempty"`
}

// OpenVPNSection holds OpenVPN server configurations.
type OpenVPNSection struct {
	Servers []OpenVPNServer `xml:"openvpn-server,omitempty"`
}

// OpenVPNServer is an OpenVPN server entry.
type OpenVPNServer struct {
	VPNID       string `xml:"vpnid"`
	Description string `xml:"description"`
	Protocol    string `xml:"protocol"`
	Port        string `xml:"local_port"`
	Tunnel      string `xml:"tunnel_network"`
	Cipher      string `xml:"data_ciphers"`
}

// WireGuardSection holds WireGuard configurations.
type WireGuardSection struct {
	Servers []WireGuardServer `xml:"server,omitempty"`
}

// WireGuardServer is a WireGuard server entry.
type WireGuardServer struct {
	Name       string `xml:"name"`
	PubKey     string `xml:"pubkey"`
	ListenPort string `xml:"listenport"`
	Tunnel     string `xml:"tunneladdress"`
}

// IPSecSection holds IPSec configurations.
type IPSecSection struct {
	// Simplified - just description for now.
	Extra []XMLExtra `xml:",any"`
}

// LoadBaseConfig reads and parses a base OPNsense config.xml file.
func LoadBaseConfig(path string) (*OpnSenseConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read base config %q: %w", path, err)
	}

	return ParseConfig(data)
}

// ParseConfig parses XML bytes into an OpnSenseConfig.
func ParseConfig(data []byte) (*OpnSenseConfig, error) {
	var cfg OpnSenseConfig
	if err := xml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config XML: %w", err)
	}

	return &cfg, nil
}

// InjectVlans adds generated VLAN data into the config.
func InjectVlans(cfg *OpnSenseConfig, vlans []generator.VlanConfig, optCounter int) {
	for i, v := range vlans {
		ifName := fmt.Sprintf("opt%d", optCounter+i)
		vlanIfName := fmt.Sprintf("vlan0.%d", v.VlanID)

		// Add VLAN entry.
		cfg.VLANs.VLANs = append(cfg.VLANs.VLANs, VLANEntry{
			If:     "igb0",
			Tag:    v.VlanID,
			Descr:  v.Description,
			VlanIf: vlanIfName,
		})

		// Add interface entry.
		gateway := netutil.GatewayIP(v.IPNetwork)
		cfg.Interfaces.Entries = append(cfg.Interfaces.Entries, InterfaceEntry{
			XMLName: xml.Name{Local: ifName},
			Enable:  "1",
			Descr:   v.Description,
			If:      vlanIfName,
			IPAddr:  gateway.String(),
			Subnet:  "24",
		})
	}
}

// InjectDHCP adds generated DHCP configurations into the config.
func InjectDHCP(cfg *OpnSenseConfig, vlans []generator.VlanConfig, dhcpConfigs []generator.DhcpServerConfig, optCounter int) {
	for i, dhcp := range dhcpConfigs {
		ifName := fmt.Sprintf("opt%d", optCounter+i)
		cfg.Dhcpd.Entries = append(cfg.Dhcpd.Entries, DhcpdEntry{
			XMLName: xml.Name{Local: ifName},
			Enable:  "1",
			Range:   DHCPRange{From: dhcp.RangeStart.String(), To: dhcp.RangeEnd.String()},
			Gateway: dhcp.Gateway.String(),
			Domain:  dhcp.DomainName,
			DNS1:    strings.Join(dhcp.DNSServers, ","),
			Lease:   fmt.Sprintf("%d", dhcp.LeaseTime),
			MaxLease: fmt.Sprintf("%d", dhcp.MaxLeaseTime),
		})
	}
}

// InjectFirewallRules adds generated firewall rules into the config.
func InjectFirewallRules(cfg *OpnSenseConfig, rules []generator.FirewallRule) {
	for _, r := range rules {
		src := RuleSrc{}
		if r.Source == "any" {
			src.Any = ""
		} else {
			src.Network = r.Source
		}

		dst := RuleDst{}
		if r.Destination == "any" {
			dst.Any = ""
		} else {
			dst.Network = r.Destination
		}
		if r.Ports != "any" {
			dst.Port = r.Ports
		}

		logStr := ""
		if r.Log {
			logStr = "1"
		}

		cfg.Filter.Rules = append(cfg.Filter.Rules, FilterRule{
			Type:        r.Action,
			Descr:       r.Description,
			Interface:   r.Interface,
			IPProtocol:  "inet",
			Protocol:    r.Protocol,
			Source:      src,
			Destination: dst,
			Log:         logStr,
			Direction:   r.Direction,
			Tracker:     fmt.Sprintf("%d", r.Tracker),
		})
	}
}

// MarshalConfig writes the config to XML with proper formatting.
func MarshalConfig(cfg *OpnSenseConfig, w io.Writer) error {
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
