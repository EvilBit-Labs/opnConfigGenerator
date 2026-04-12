package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestRootCmd creates a fresh command tree to avoid state leakage between tests.
// Each test gets its own root command with all subcommands attached.
//
// NOTE: Because cobra flag bindings point to package-level variables (quiet, noColor,
// output, format, etc.), tests that call this function must NOT run in parallel.
func newTestRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:     "opnConfigGenerator",
		Short:   "Generate realistic OPNsense configuration files with fake data",
		Version: "test",
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			setupLogging()
		},
	}

	root.PersistentFlags().BoolVar(&quiet, "quiet", false, "suppress output except errors")
	root.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	root.PersistentFlags().StringVarP(&output, "output", "o", "", "output file path (default: stdout)")

	genCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate OPNsense configuration data",
		RunE:  runGenerate,
	}
	genCmd.Flags().StringVar(&format, "format", "", "output format (csv|xml)")
	if err := genCmd.MarkFlagRequired("format"); err != nil {
		panic(err)
	}
	genCmd.Flags().IntVarP(&count, "count", "c", 10, "number of VLANs to generate (1-10000)")
	genCmd.Flags().StringVar(&baseConfig, "base-config", "", "base OPNsense XML template")
	genCmd.Flags().StringVar(&csvFile, "csv-file", "", "read VLANs from existing CSV file")
	genCmd.Flags().IntVar(&firewallNr, "firewall-nr", 1, "firewall instance number")
	genCmd.Flags().IntVar(&optCounter, "opt-counter", 6, "starting interface counter")
	genCmd.Flags().BoolVar(&force, "force", false, "overwrite existing output files")
	genCmd.Flags().Int64Var(&seed, "seed", 0, "RNG seed for reproducibility")
	genCmd.Flags().BoolVar(&includeFirewallRules, "include-firewall-rules", false, "generate firewall rules")
	genCmd.Flags().StringVar(&firewallRuleComplexity, "firewall-rule-complexity", "basic", "complexity")
	genCmd.Flags().StringVar(&vlanRange, "vlan-range", "", "VLAN range spec")
	genCmd.Flags().IntVar(&vpnCount, "vpn-count", 0, "number of VPN configurations")
	genCmd.Flags().IntVar(&natMappings, "nat-mappings", 0, "number of NAT rules")
	genCmd.Flags().StringVar(&wanAssignments, "wan-assignments", "single", "WAN strategy")

	valCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate OPNsense configuration files",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errors.New("validate command not yet implemented")
		},
	}
	valCmd.Flags().StringVarP(&inputFile, "input", "i", "", "input file to validate")
	if err := valCmd.MarkFlagRequired("input"); err != nil {
		panic(err)
	}
	valCmd.Flags().StringVar(&inputFormat, "format", "", "input format")
	valCmd.Flags().IntVar(&maxErrors, "max-errors", 10, "maximum number of errors to report")

	compCmd := &cobra.Command{
		Use:       "completion [bash|zsh|fish|powershell]",
		Short:     "Generate completion script",
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(c *cobra.Command, args []string) error {
			out := c.OutOrStdout()
			switch args[0] {
			case "bash":
				return c.Root().GenBashCompletion(out)
			case "zsh":
				return c.Root().GenZshCompletion(out)
			case "fish":
				return c.Root().GenFishCompletion(out, true)
			case "powershell":
				return c.Root().GenPowerShellCompletionWithDesc(out)
			default:
				return nil
			}
		},
	}

	root.AddCommand(genCmd, valCmd, compCmd)
	return root
}

// executeCommand runs the command tree with the given args and captures stdout/stderr.
func executeCommand(root *cobra.Command, args ...string) (string, error) {
	stdoutBuf := new(bytes.Buffer)
	root.SetOut(stdoutBuf)
	root.SetErr(new(bytes.Buffer))
	root.SetArgs(args)
	err := root.Execute()
	return stdoutBuf.String(), err
}

// baseConfigPath returns the absolute path to the base-config.xml test fixture.
func baseConfigPath(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs(filepath.Join("..", "testdata", "base-config.xml"))
	require.NoError(t, err)
	return abs
}

// --- Root Command Tests ---

func TestRootHelp(t *testing.T) {
	cmd := newTestRootCmd()
	stdout, err := executeCommand(cmd, "--help")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Generate realistic OPNsense configuration files")
	assert.Contains(t, stdout, "generate")
	assert.Contains(t, stdout, "validate")
	assert.Contains(t, stdout, "completion")
}

func TestRootVersion(t *testing.T) {
	cmd := newTestRootCmd()
	stdout, err := executeCommand(cmd, "--version")
	require.NoError(t, err)
	assert.Contains(t, stdout, "test")
}

// --- Generate Command Tests ---

func TestGenerateMissingFormat(t *testing.T) {
	cmd := newTestRootCmd()
	_, err := executeCommand(cmd, "generate")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "format")
}

func TestGenerateInvalidFormat(t *testing.T) {
	cmd := newTestRootCmd()
	_, err := executeCommand(cmd, "generate", "--format", "yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid format")
}

func TestGenerateCSVToFile(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "vlans.csv")

	cmd := newTestRootCmd()
	_, err := executeCommand(cmd, "generate", "--format", "csv", "--count", "5", "--seed", "42", "--output", outPath)
	require.NoError(t, err)

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	content := string(data)
	// CSV uses German headers with UTF-8 BOM prefix.
	lines := strings.Split(strings.TrimSpace(content), "\n")
	assert.Len(t, lines, 6, "expected header + 5 data rows")
	assert.Contains(t, lines[0], "VLAN")
}

func TestGenerateXMLRequiresBaseConfig(t *testing.T) {
	cmd := newTestRootCmd()
	_, err := executeCommand(cmd, "generate", "--format", "xml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--base-config is required")
}

func TestGenerateXMLWithBaseConfig(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "output.xml")
	baseConfigPath := baseConfigPath(t)

	cmd := newTestRootCmd()
	_, err := executeCommand(cmd,
		"generate", "--format", "xml",
		"--count", "3",
		"--base-config", baseConfigPath,
		"--seed", "42",
		"--output", outPath,
	)
	require.NoError(t, err)

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "<?xml")
	assert.Contains(t, content, "<opnsense>")
	assert.Contains(t, content, "<vlans>")
}

func TestGenerateCSVThreeRows(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "vlans.csv")

	cmd := newTestRootCmd()
	_, err := executeCommand(cmd, "generate", "--format", "csv", "--count", "3", "--seed", "42", "--output", outPath)
	require.NoError(t, err)

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	assert.Len(t, lines, 4, "expected header + 3 data rows")
}

func TestGenerateDeterministicSeed(t *testing.T) {
	tmpDir := t.TempDir()
	outPath1 := filepath.Join(tmpDir, "run1.csv")
	outPath2 := filepath.Join(tmpDir, "run2.csv")

	cmd1 := newTestRootCmd()
	_, err := executeCommand(
		cmd1,
		"generate",
		"--format",
		"csv",
		"--count",
		"5",
		"--seed",
		"42",
		"--output",
		outPath1,
	)
	require.NoError(t, err)

	cmd2 := newTestRootCmd()
	_, err = executeCommand(
		cmd2,
		"generate",
		"--format",
		"csv",
		"--count",
		"5",
		"--seed",
		"42",
		"--output",
		outPath2,
	)
	require.NoError(t, err)

	data1, err := os.ReadFile(outPath1)
	require.NoError(t, err)
	data2, err := os.ReadFile(outPath2)
	require.NoError(t, err)
	assert.Equal(t, string(data1), string(data2), "same seed should produce identical output")
}

func TestGenerateFileExistsWithoutForce(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "existing.csv")

	require.NoError(t, os.WriteFile(outPath, []byte("existing"), 0o600))

	cmd := newTestRootCmd()
	_, err := executeCommand(cmd, "generate", "--format", "csv", "--count", "3", "--seed", "42", "--output", outPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestGenerateFileExistsWithForce(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "overwrite.csv")

	require.NoError(t, os.WriteFile(outPath, []byte("old"), 0o600))

	cmd := newTestRootCmd()
	_, err := executeCommand(
		cmd,
		"generate",
		"--format",
		"csv",
		"--count",
		"3",
		"--seed",
		"42",
		"--output",
		outPath,
		"--force",
	)
	require.NoError(t, err)

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "VLAN", "file should be overwritten with CSV data")
}

func TestGenerateXMLWithFirewallRules(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "fw.xml")
	baseConfigPath := baseConfigPath(t)

	cmd := newTestRootCmd()
	_, err := executeCommand(cmd,
		"generate", "--format", "xml",
		"--count", "3",
		"--base-config", baseConfigPath,
		"--seed", "42",
		"--include-firewall-rules",
		"--output", outPath,
	)
	require.NoError(t, err)

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "<opnsense>")
}

func TestGenerateXMLRejectsNatMappings(t *testing.T) {
	baseConfigPath := baseConfigPath(t)

	cmd := newTestRootCmd()
	_, err := executeCommand(cmd,
		"generate", "--format", "xml",
		"--count", "3",
		"--base-config", baseConfigPath,
		"--nat-mappings", "5",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not yet supported")
}

func TestGenerateXMLRejectsVpnCount(t *testing.T) {
	baseConfigPath := baseConfigPath(t)

	cmd := newTestRootCmd()
	_, err := executeCommand(cmd,
		"generate", "--format", "xml",
		"--count", "3",
		"--base-config", baseConfigPath,
		"--vpn-count", "2",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not yet supported")
}

func TestGenerateInvalidCount(t *testing.T) {
	cmd := newTestRootCmd()
	_, err := executeCommand(cmd, "generate", "--format", "csv", "--count", "0")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--count must be between")
}

// --- Validate Command Tests ---

func TestValidateNotImplemented(t *testing.T) {
	cmd := newTestRootCmd()
	_, err := executeCommand(cmd, "validate", "--input", "dummy.xml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not yet implemented")
}

func TestValidateMissingInput(t *testing.T) {
	cmd := newTestRootCmd()
	_, err := executeCommand(cmd, "validate")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "input")
}

// --- Completion Command Tests ---

func TestCompletionBash(t *testing.T) {
	cmd := newTestRootCmd()
	stdout, err := executeCommand(cmd, "completion", "bash")
	require.NoError(t, err)
	assert.Contains(t, stdout, "bash completion", "should contain bash completion markers")
}

func TestCompletionZsh(t *testing.T) {
	cmd := newTestRootCmd()
	stdout, err := executeCommand(cmd, "completion", "zsh")
	require.NoError(t, err)
	assert.NotEmpty(t, stdout, "zsh completion should produce output")
}

func TestCompletionFish(t *testing.T) {
	cmd := newTestRootCmd()
	stdout, err := executeCommand(cmd, "completion", "fish")
	require.NoError(t, err)
	assert.Contains(t, stdout, "fish", "should contain fish completion content")
}

func TestCompletionInvalidShell(t *testing.T) {
	cmd := newTestRootCmd()
	_, err := executeCommand(cmd, "completion", "invalid")
	require.Error(t, err)
}

func TestCompletionNoArgs(t *testing.T) {
	cmd := newTestRootCmd()
	_, err := executeCommand(cmd, "completion")
	require.Error(t, err)
}

// --- normalizeStringFlag Tests ---

func TestNormalizeStringFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"lowercase", "CSV", "csv"},
		{"trim whitespace", "  xml  ", "xml"},
		{"mixed", "  XML  ", "xml"},
		{"already normalized", "csv", "csv"},
		{"empty string", "", ""},
		{"tabs and spaces", "\t csv \t", "csv"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := normalizeStringFlag(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

// --- getOutputWriter Tests ---

func TestGetOutputWriterStdout(t *testing.T) {
	origOutput := output
	output = ""
	t.Cleanup(func() { output = origOutput })

	w, needClose, err := getOutputWriter()
	require.NoError(t, err)
	assert.False(t, needClose, "stdout should not need closing")
	assert.Equal(t, os.Stdout, w, "should return os.Stdout when output is empty")
}

func TestGetOutputWriterCreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "test-output.txt")

	origOutput := output
	origForce := force
	output = outPath
	force = false
	t.Cleanup(func() {
		output = origOutput
		force = origForce
	})

	w, needClose, err := getOutputWriter()
	require.NoError(t, err)
	assert.True(t, needClose, "file should need closing")
	require.NotNil(t, w)
	require.NoError(t, w.Close())

	_, statErr := os.Stat(outPath)
	assert.NoError(t, statErr, "file should have been created")
}

func TestGetOutputWriterExistingFileNoForce(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "existing.txt")
	require.NoError(t, os.WriteFile(outPath, []byte("data"), 0o600))

	origOutput := output
	origForce := force
	output = outPath
	force = false
	t.Cleanup(func() {
		output = origOutput
		force = origForce
	})

	_, _, err := getOutputWriter()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestGetOutputWriterExistingFileWithForce(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "existing.txt")
	require.NoError(t, os.WriteFile(outPath, []byte("data"), 0o600))

	origOutput := output
	origForce := force
	output = outPath
	force = true
	t.Cleanup(func() {
		output = origOutput
		force = origForce
	})

	w, needClose, err := getOutputWriter()
	require.NoError(t, err)
	assert.True(t, needClose, "file should need closing")
	require.NotNil(t, w)
	require.NoError(t, w.Close())
}

// --- Execute Tests ---

func TestExecuteSetsVersion(t *testing.T) {
	origVersion := rootCmd.Version
	t.Cleanup(func() { rootCmd.Version = origVersion })

	rootCmd.SetArgs([]string{"--version"})
	stdoutBuf := new(bytes.Buffer)
	rootCmd.SetOut(stdoutBuf)

	err := Execute("1.2.3-test")
	require.NoError(t, err)
	assert.Contains(t, stdoutBuf.String(), "1.2.3-test")
}
