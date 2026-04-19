package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var (
	quiet   bool
	noColor bool
	output  string
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "opnConfigGenerator",
	Short: "Generate realistic OPNsense configuration files with fake data",
	Long: `opnConfigGenerator is the reverse serializer for opnDossier's *model.CommonDevice.
It produces a valid OPNsense config.xml from a synthetic CommonDevice populated by a faker.

Phase 1 coverage:
  • System: hostname, domain, timezone, DNS/NTP servers
  • Interfaces: WAN (DHCP), LAN (static RFC 1918), per-VLAN opt interfaces
  • VLANs with unique 802.1Q tags on a shared physical parent
  • DHCP scopes per statically-addressed interface (ISC DHCP)
  • Default allow firewall rules per non-WAN interface (opt-in)

Deferred to follow-up plans: NAT, VPN (OpenVPN/WireGuard/IPsec), Users/Groups,
Certificates, IDS, HighAvailability, VirtualIPs, Bridges, GIF/GRE/LAGG, PPP,
CaptivePortal, Kea DHCP, Monit, Netflow, TrafficShaper, pfSense target.

Zero arguments emits a valid config.xml on stdout. With --base-config, the
serializer overlays generated content onto an existing document, preserving
fields Phase 1 does not own.`,
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		// Set up logging based on flags and environment
		setupLogging()
	},
}

// Execute runs the root command with the given version string.
func Execute(version string) error {
	rootCmd.Version = version
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&quiet, "quiet", false, "suppress output except errors")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "output file path (default: stdout)")

	// Add subcommands
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(completionCmd)
}

func setupLogging() {
	// Check environment variables
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		noColor = true
	}

	// Configure logger
	logger := log.New(os.Stderr)

	if quiet {
		logger.SetLevel(log.ErrorLevel)
	} else {
		logger.SetLevel(log.InfoLevel)
	}

	if noColor {
		logger.SetColorProfile(0)
	}

	log.SetDefault(logger)
}

// getOutputWriter returns the appropriate output destination.
// If an output file is specified and --force is not set, it uses O_EXCL for atomic creation.
func getOutputWriter() (*os.File, bool, error) {
	if output == "" {
		return os.Stdout, false, nil
	}

	if force {
		file, err := os.Create(output)
		if err != nil {
			return nil, false, fmt.Errorf("create output file %s: %w", output, err)
		}
		return file, true, nil
	}

	// Use O_EXCL for atomic "create if not exists" — no TOCTOU race.
	//nolint:gosec // Output file path is user-specified CLI input
	file, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return nil, false, fmt.Errorf("output file %q already exists (use --force to overwrite)", output)
		}
		return nil, false, fmt.Errorf("create output file %s: %w", output, err)
	}

	return file, true, nil
}

// normalizeStringFlag normalizes string flags by trimming whitespace and converting to lowercase.
func normalizeStringFlag(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}
