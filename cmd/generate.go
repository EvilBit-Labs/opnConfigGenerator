package cmd

import (
	"fmt"

	"github.com/EvilBit-Labs/opnConfigGenerator/internal/csvio"
	"github.com/EvilBit-Labs/opnConfigGenerator/internal/faker"
	"github.com/EvilBit-Labs/opnConfigGenerator/internal/opnsensegen"
	serializer "github.com/EvilBit-Labs/opnConfigGenerator/internal/serializer/opnsense"
	"github.com/EvilBit-Labs/opnDossier/pkg/model"
	"github.com/EvilBit-Labs/opnDossier/pkg/schema/opnsense"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

const (
	formatXML = "xml"
	formatCSV = "csv"

	// defaultVlanCount is the number of VLANs generated when no count is
	// supplied.
	defaultVlanCount = 10
)

// maxVlanCount mirrors faker.MaxVLANCount so the CLI and library bound
// validation use the same number.
var maxVlanCount = faker.MaxVLANCount

var (
	outputFormat     string
	vlanCount        int
	baseConfigPath   string
	includeFirewall  bool
	seed             int64
	force            bool
	hostnameOverride string
	domainOverride   string
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate OPNsense configuration data",
	Long: `Generate realistic OPNsense configuration data.

With no arguments, emits a valid OPNsense config.xml on stdout.

Examples:
  # Zero-input: valid config.xml with 10 VLANs on stdout
  opnConfigGenerator generate

  # Reproducible output
  opnConfigGenerator generate --seed 42

  # With default firewall rules and 20 VLANs
  opnConfigGenerator generate --vlan-count 20 --firewall-rules

  # Overlay generated content onto an existing config
  opnConfigGenerator generate --base-config existing.xml

  # CSV inspection dump
  opnConfigGenerator generate --format csv`,
	RunE: runGenerate,
}

func init() {
	generateCmd.Flags().StringVar(&outputFormat, "format", formatXML, "output format (xml|csv)")
	generateCmd.Flags().IntVarP(&vlanCount, "vlan-count", "n", defaultVlanCount, "number of VLANs to generate (0-4093)")
	generateCmd.Flags().
		StringVar(&baseConfigPath, "base-config", "", "optional base OPNsense config.xml to overlay generated content onto")
	generateCmd.Flags().
		BoolVar(&includeFirewall, "firewall-rules", false, "include default allow-all-to-any rules per interface")
	generateCmd.Flags().Int64Var(&seed, "seed", 0, "RNG seed for reproducibility (0 = random)")
	generateCmd.Flags().BoolVar(&force, "force", false, "overwrite existing output file")
	generateCmd.Flags().StringVar(&hostnameOverride, "hostname", "", "override the generated hostname")
	generateCmd.Flags().StringVar(&domainOverride, "domain", "", "override the generated domain")
}

// buildOpnSenseDocument routes the serialize step. When --base-config is set
// we go through Overlay (which serializes the device against the loaded base);
// otherwise we call Serialize directly. Keeping the branch here means the XML
// path never serializes the device twice.
func buildOpnSenseDocument(device *model.CommonDevice) (*opnsense.OpnSenseDocument, error) {
	if baseConfigPath == "" {
		doc, err := serializer.Serialize(device)
		if err != nil {
			return nil, fmt.Errorf("serialize: %w", err)
		}
		return doc, nil
	}
	base, err := opnsensegen.LoadBaseConfig(baseConfigPath)
	if err != nil {
		return nil, fmt.Errorf("load base config: %w", err)
	}
	doc, err := serializer.Overlay(base, device)
	if err != nil {
		return nil, fmt.Errorf("overlay: %w", err)
	}
	return doc, nil
}

func runGenerate(_ *cobra.Command, _ []string) (err error) {
	format := normalizeStringFlag(outputFormat)
	switch format {
	case formatXML, formatCSV:
	default:
		return fmt.Errorf("invalid format %q: must be xml or csv", outputFormat)
	}

	if vlanCount < 0 || vlanCount > maxVlanCount {
		return fmt.Errorf("--vlan-count must be between 0 and %d, got %d", maxVlanCount, vlanCount)
	}

	if format != formatXML && baseConfigPath != "" {
		return fmt.Errorf("--base-config is only supported with --format %s", formatXML)
	}

	device, err := faker.NewCommonDevice(
		faker.WithSeed(seed),
		faker.WithVLANCount(vlanCount),
		faker.WithFirewallRules(includeFirewall),
		faker.WithHostname(hostnameOverride),
		faker.WithDomain(domainOverride),
		// Today the CLI always targets OPNsense; when the pfSense
		// serializer lands, route this by a user-facing flag.
		faker.WithDeviceType(model.DeviceTypeOPNsense),
	)
	if err != nil {
		return fmt.Errorf("generate device: %w", err)
	}

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

	switch format {
	case formatCSV:
		if cerr := csvio.WriteVlanCSV(w, device); cerr != nil {
			return fmt.Errorf("write CSV: %w", cerr)
		}
		log.Info("wrote CSV output", "vlans", len(device.VLANs))
		return nil
	case formatXML:
		doc, sErr := buildOpnSenseDocument(device)
		if sErr != nil {
			return sErr
		}
		if mErr := opnsensegen.MarshalConfig(doc, w); mErr != nil {
			return fmt.Errorf("write XML: %w", mErr)
		}
		log.Info("wrote XML output", "vlans", len(device.VLANs), "interfaces", len(device.Interfaces))
		return nil
	}
	return fmt.Errorf("unsupported format: %s", format)
}
