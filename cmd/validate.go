package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	inputFile   string
	inputFormat string
	maxErrors   int
)

// validateCmd represents the validate command.
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate OPNsense configuration files",
	Long: `Validate OPNsense configuration files for correctness and compliance.

This command validates OPNsense configuration files against the OPNsense schema
and checks for common configuration errors, conflicts, and best practices.

The validator can detect:
  • XML schema violations
  • Network configuration conflicts (IP overlaps, VLAN conflicts)
  • Invalid interface assignments
  • Malformed firewall rules
  • Missing required dependencies
  • Security misconfigurations

Examples:
  # Validate an OPNsense config.xml file
  opnConfigGenerator validate --input config.xml

  # Validate with format auto-detection
  opnConfigGenerator validate --input network-config.xml --format xml

  # Limit error reporting
  opnConfigGenerator validate --input config.xml --max-errors 5`,
	RunE: func(_ *cobra.Command, _ []string) error {
		return errors.New("validate command not yet implemented")
	},
}

func init() {
	// Required flags
	validateCmd.Flags().StringVarP(&inputFile, "input", "i", "", "input file to validate")
	if err := validateCmd.MarkFlagRequired("input"); err != nil {
		panic(fmt.Sprintf("failed to mark input flag required: %v", err))
	}

	// Optional flags
	validateCmd.Flags().StringVar(&inputFormat, "format", "", "input format (auto-detect if not specified)")
	//nolint:mnd // CLI flag default value
	validateCmd.Flags().IntVar(&maxErrors, "max-errors", 10, "maximum number of errors to report")
}
