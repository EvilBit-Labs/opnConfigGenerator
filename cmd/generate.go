package cmd

import (
	"errors"
	"fmt"
	"math/rand/v2"

	"github.com/EvilBit-Labs/opnConfigGenerator/internal/csvio"
	"github.com/EvilBit-Labs/opnConfigGenerator/internal/generator"
	"github.com/EvilBit-Labs/opnConfigGenerator/internal/opnsensegen"
	"github.com/EvilBit-Labs/opnConfigGenerator/internal/validate"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

const (
	// formatXML is the XML output format identifier.
	formatXML = "xml"
	// formatCSV is the CSV output format identifier.
	formatCSV = "csv"
)

var (
	format                 string
	count                  int
	baseConfig             string
	csvFile                string
	firewallNr             int
	optCounter             int
	force                  bool
	seed                   int64
	includeFirewallRules   bool
	firewallRuleComplexity string
	vlanRange              string
	vpnCount               int
	natMappings            int
	wanAssignments         string
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate OPNsense configuration data",
	Long: `Generate realistic OPNsense configuration data in various formats.

Examples:
  # Generate 25 VLANs in XML format
  opnConfigGenerator generate --format xml --count 25 --base-config config.xml

  # Generate with firewall rules
  opnConfigGenerator generate --format xml --count 10 --include-firewall-rules

  # Generate CSV data
  opnConfigGenerator generate --format csv --count 50 --output network-data.csv

  # Generate with VPN and NAT (CSV only — XML serialization pending)
  opnConfigGenerator generate --format csv --count 15 --vpn-count 3 --nat-mappings 10`,
	RunE: runGenerate,
}

func init() {
	generateCmd.Flags().StringVar(&format, "format", "", "output format (csv|xml)")
	if err := generateCmd.MarkFlagRequired("format"); err != nil {
		panic(fmt.Sprintf("failed to mark format flag required: %v", err))
	}

	//nolint:mnd // CLI flag default value
	generateCmd.Flags().IntVarP(&count, "count", "c", 10, "number of VLANs to generate (1-4085)")
	generateCmd.Flags().
		StringVar(&baseConfig, "base-config", "", "base OPNsense XML template (required for xml format)")
	generateCmd.Flags().StringVar(&csvFile, "csv-file", "", "read VLANs from existing CSV file")

	generateCmd.Flags().IntVar(&firewallNr, "firewall-nr", 1, "firewall instance number (1-999)")
	//nolint:mnd // CLI flag default value
	generateCmd.Flags().IntVar(&optCounter, "opt-counter", 6, "starting interface counter")

	generateCmd.Flags().BoolVar(&force, "force", false, "overwrite existing output files")
	generateCmd.Flags().Int64Var(&seed, "seed", 0, "RNG seed for reproducibility (0 = random)")

	generateCmd.Flags().BoolVar(&includeFirewallRules, "include-firewall-rules", false, "generate firewall rules")
	generateCmd.Flags().
		StringVar(&firewallRuleComplexity, "firewall-rule-complexity", "basic", "complexity (basic|intermediate|advanced)")

	generateCmd.Flags().StringVar(&vlanRange, "vlan-range", "", "VLAN range spec (e.g., '100-150,200-250')")
	generateCmd.Flags().IntVar(&vpnCount, "vpn-count", 0, "number of VPN configurations")
	generateCmd.Flags().IntVar(&natMappings, "nat-mappings", 0, "number of NAT rules")
	generateCmd.Flags().StringVar(&wanAssignments, "wan-assignments", "single", "WAN strategy (single|multi|balanced)")
}

// CLI validation constants.
const (
	maxFirewallNr = 999
	minOptCounter = 0
)

func runGenerate(_ *cobra.Command, _ []string) error {
	normalizedFormat := normalizeStringFlag(format)

	// Validate format.
	switch normalizedFormat {
	case formatCSV, formatXML:
		// Valid.
	default:
		return fmt.Errorf("invalid format %q: must be csv or xml", format)
	}

	// XML format requires base config.
	if normalizedFormat == formatXML && baseConfig == "" {
		return errors.New("--base-config is required for xml format")
	}

	// Validate count range.
	if count <= 0 || count > generator.MaxUniqueVlans {
		return fmt.Errorf("--count must be between 1 and %d, got %d", generator.MaxUniqueVlans, count)
	}

	// Validate firewallNr range.
	if firewallNr < 1 || firewallNr > maxFirewallNr {
		return fmt.Errorf("--firewall-nr must be between 1 and %d, got %d", maxFirewallNr, firewallNr)
	}

	// Validate optCounter range.
	if optCounter < minOptCounter {
		return fmt.Errorf("--opt-counter must be non-negative, got %d", optCounter)
	}

	// NAT and VPN injection into XML is not yet implemented.
	if normalizedFormat == formatXML && (natMappings > 0 || vpnCount > 0) {
		return errors.New("--nat-mappings and --vpn-count are not yet supported for XML output")
	}

	// Parse WAN assignment strategy.
	wanStrategy, err := generator.ParseWanAssignment(normalizeStringFlag(wanAssignments))
	if err != nil {
		return err
	}

	// Parse firewall complexity.
	complexity, err := generator.ParseFirewallComplexity(normalizeStringFlag(firewallRuleComplexity))
	if err != nil {
		return err
	}

	// Set up seed.
	var seedPtr *int64
	if seed != 0 {
		seedPtr = &seed
	}

	log.Info("generating configuration", "format", normalizedFormat, "count", count)

	// Generate VLANs.
	vlanGen := generator.NewVlanGenerator(seedPtr, wanStrategy)
	vlans, err := vlanGen.GenerateBatch(count)
	if err != nil {
		return fmt.Errorf("generate VLANs: %w", err)
	}

	// Validate VLANs.
	result := validate.Vlans(vlans)
	if !result.IsValid() {
		return result.Error()
	}

	log.Info("generated VLANs", "count", len(vlans))

	// Generate firewall rules if requested.
	var fwRules []generator.FirewallRule
	if includeFirewallRules {
		fwGen := generator.NewFirewallGenerator(seedPtr)
		fwRules = fwGen.GenerateRulesForBatch(vlans, complexity)
		log.Info("generated firewall rules", "count", len(fwRules))
	}

	// Output based on format.
	switch normalizedFormat {
	case formatCSV:
		return outputCSV(vlans)
	case formatXML:
		return outputXML(vlans, fwRules, seedPtr)
	default:
		return fmt.Errorf("unsupported format: %s", normalizedFormat)
	}
}

func outputCSV(vlans []generator.VlanConfig) (err error) {
	w, needClose, err := getOutputWriter()
	if err != nil {
		return err
	}
	if needClose {
		defer func() {
			if cerr := w.Close(); cerr != nil && err == nil {
				err = fmt.Errorf("close output file: %w", cerr)
			}
		}()
	}

	if err := csvio.WriteVlanCSV(w, vlans); err != nil {
		return fmt.Errorf("write CSV: %w", err)
	}

	log.Info("wrote CSV output", "vlans", len(vlans))
	return nil
}

func outputXML(
	vlans []generator.VlanConfig,
	fwRules []generator.FirewallRule,
	seedPtr *int64,
) (err error) {
	// Load base config.
	cfg, err := opnsensegen.LoadBaseConfig(baseConfig)
	if err != nil {
		return fmt.Errorf("load base config: %w", err)
	}

	// Inject generated data.
	opnsensegen.InjectVlans(cfg, vlans, optCounter)

	// Generate and inject DHCP configs using same seed logic as other generators.
	var rng *rand.Rand
	if seedPtr != nil {
		//nolint:gosec // Deterministic fake data generation, not security-sensitive
		rng = rand.New(rand.NewPCG(uint64(*seedPtr), 0))
	} else {
		//nolint:gosec // Deterministic fake data generation, not security-sensitive
		rng = rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	}
	dhcpConfigs := make([]generator.DhcpServerConfig, len(vlans))
	for i, v := range vlans {
		dhcpConfigs[i] = generator.DeriveDHCPConfig(v, rng)
	}
	opnsensegen.InjectDHCP(cfg, dhcpConfigs, optCounter)

	// Inject firewall rules.
	if len(fwRules) > 0 {
		opnsensegen.InjectFirewallRules(cfg, fwRules)
	}

	// Get output writer.
	w, needClose, err := getOutputWriter()
	if err != nil {
		return err
	}
	if needClose {
		defer func() {
			if cerr := w.Close(); cerr != nil && err == nil {
				err = fmt.Errorf("close output file: %w", cerr)
			}
		}()
	}

	// Write output.
	if err := opnsensegen.MarshalConfig(cfg, w); err != nil {
		return fmt.Errorf("write XML: %w", err)
	}

	log.Info("wrote XML output", "vlans", len(vlans), "rules", len(fwRules))
	return nil
}
