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

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "opnConfigGenerator",
	Short: "Generate realistic OPNsense configuration files with fake data",
	Long: `opnConfigGenerator is a command-line tool for generating realistic OPNsense config.xml files
populated with fake but valid network configuration data. It's designed for testing, development,
and demonstration purposes where you need realistic OPNsense configurations without exposing
sensitive network information.

Features:
  • Generate realistic VLAN configurations
  • Create valid interface assignments
  • Generate DHCP pools and static mappings
  • Create firewall rules with proper dependencies
  • Generate VPN configurations (OpenVPN, WireGuard, IPSec)
  • Create NAT rules and port forwards
  • Support for various output formats (XML, CSV)
  • Configurable generation parameters`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Set up logging based on flags and environment
		setupLogging()
	},
}

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

// getOutputWriter returns the appropriate output destination
func getOutputWriter() (*os.File, bool, error) {
	if output == "" {
		return os.Stdout, false, nil
	}

	file, err := os.Create(output)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create output file %s: %w", output, err)
	}

	return file, true, nil
}

// normalizeStringFlag normalizes string flags by trimming whitespace and converting to lowercase
func normalizeStringFlag(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}