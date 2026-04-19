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
	Long: `Validate OPNsense configuration files.

Not yet implemented — this subcommand is reserved for a future phase and
currently returns an error. Flags are defined only so they are stable when
implementation lands.`,
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
