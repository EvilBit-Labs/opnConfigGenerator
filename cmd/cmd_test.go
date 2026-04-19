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
// NOTE: Because cobra flag bindings point to package-level variables, tests
// that call this function must NOT run in parallel.
func newTestRootCmd() *cobra.Command {
	// Reset package-level flag vars to their defaults before rebuilding the
	// command tree. This matters because Cobra wires flags to these variables
	// directly; state leaks across tests otherwise.
	outputFormat = formatXML
	vlanCount = defaultVlanCount
	baseConfigPath = ""
	includeFirewall = false
	seed = 0
	force = false
	hostnameOverride = ""
	domainOverride = ""
	output = ""
	quiet = false
	noColor = false
	// validate subcommand globals (defined in cmd/validate.go).
	inputFile = ""
	inputFormat = ""
	maxErrors = 10

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
	genCmd.Flags().StringVar(&outputFormat, "format", formatXML, "output format (xml|csv)")
	genCmd.Flags().IntVarP(&vlanCount, "vlan-count", "n", defaultVlanCount, "number of VLANs to generate (0-4093)")
	genCmd.Flags().StringVar(&baseConfigPath, "base-config", "", "optional base OPNsense config.xml")
	genCmd.Flags().BoolVar(&includeFirewall, "firewall-rules", false, "include default firewall rules")
	genCmd.Flags().Int64Var(&seed, "seed", 0, "RNG seed for reproducibility")
	genCmd.Flags().BoolVar(&force, "force", false, "overwrite existing output file")
	genCmd.Flags().StringVar(&hostnameOverride, "hostname", "", "override the generated hostname")
	genCmd.Flags().StringVar(&domainOverride, "domain", "", "override the generated domain")

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

// baseConfigFixture returns the absolute path to the base-config.xml fixture.
func baseConfigFixture(t *testing.T) string {
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

func TestGenerateZeroArgsProducesXML(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "default.xml")

	cmd := newTestRootCmd()
	_, err := executeCommand(cmd, "generate", "--seed", "1", "--output", outPath)
	require.NoError(t, err)

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "<?xml")
	assert.Contains(t, content, "<opnsense>")
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
	_, err := executeCommand(cmd, "generate",
		"--format", "csv",
		"--vlan-count", "5",
		"--seed", "42",
		"--output", outPath,
	)
	require.NoError(t, err)

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	content := string(data)
	lines := strings.Split(strings.TrimSpace(content), "\n")
	assert.Len(t, lines, 6, "header + 5 data rows")
	assert.Contains(t, lines[0], "VLAN")
}

func TestGenerateXMLWithoutBaseConfig(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "scratch.xml")

	cmd := newTestRootCmd()
	_, err := executeCommand(cmd,
		"generate",
		"--vlan-count", "3",
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

func TestGenerateXMLWithBaseConfig(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "overlay.xml")
	baseCfgPath := baseConfigFixture(t)

	cmd := newTestRootCmd()
	_, err := executeCommand(cmd,
		"generate",
		"--vlan-count", "3",
		"--base-config", baseCfgPath,
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

func TestGenerateBaseConfigMissing(t *testing.T) {
	cmd := newTestRootCmd()
	_, err := executeCommand(cmd, "generate", "--base-config", "/does/not/exist.xml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "load base config")
}

func TestGenerateDeterministicSeed(t *testing.T) {
	tmpDir := t.TempDir()
	outPath1 := filepath.Join(tmpDir, "run1.xml")
	outPath2 := filepath.Join(tmpDir, "run2.xml")

	cmd1 := newTestRootCmd()
	_, err := executeCommand(cmd1, "generate",
		"--vlan-count", "5",
		"--seed", "42",
		"--output", outPath1,
	)
	require.NoError(t, err)

	cmd2 := newTestRootCmd()
	_, err = executeCommand(cmd2, "generate",
		"--vlan-count", "5",
		"--seed", "42",
		"--output", outPath2,
	)
	require.NoError(t, err)

	data1, err := os.ReadFile(outPath1)
	require.NoError(t, err)
	data2, err := os.ReadFile(outPath2)
	require.NoError(t, err)
	assert.Equal(t, string(data1), string(data2), "same seed must produce byte-identical output")
}

func TestGenerateFileExistsWithoutForce(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "existing.xml")
	require.NoError(t, os.WriteFile(outPath, []byte("existing"), 0o600))

	cmd := newTestRootCmd()
	_, err := executeCommand(cmd, "generate", "--seed", "42", "--output", outPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestGenerateFileExistsWithForce(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "overwrite.xml")
	require.NoError(t, os.WriteFile(outPath, []byte("old"), 0o600))

	cmd := newTestRootCmd()
	_, err := executeCommand(cmd, "generate",
		"--seed", "42",
		"--output", outPath,
		"--force",
	)
	require.NoError(t, err)

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "<opnsense>", "file must be overwritten with generated XML")
}

func TestGenerateXMLWithFirewallRules(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "fw.xml")

	cmd := newTestRootCmd()
	_, err := executeCommand(cmd, "generate",
		"--vlan-count", "3",
		"--seed", "42",
		"--firewall-rules",
		"--output", outPath,
	)
	require.NoError(t, err)

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "<rule>")
}

func TestGenerateInvalidVlanCount(t *testing.T) {
	cmd := newTestRootCmd()
	_, err := executeCommand(cmd, "generate", "--vlan-count", "-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--vlan-count")
}

func TestGenerateVlanCountZero(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "zero.xml")

	cmd := newTestRootCmd()
	_, err := executeCommand(cmd, "generate",
		"--vlan-count", "0",
		"--seed", "1",
		"--output", outPath,
	)
	require.NoError(t, err)

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "<opnsense>")
	// No <vlan> children in the <vlans> section.
	assert.NotContains(t, content, "<vlan>")
}

func TestGenerateVlanCountOne(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "one.xml")

	cmd := newTestRootCmd()
	_, err := executeCommand(cmd, "generate",
		"--vlan-count", "1",
		"--seed", "1",
		"--output", outPath,
	)
	require.NoError(t, err)

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	assert.Equal(t, 1, strings.Count(string(data), "<vlan>"))
}

func TestGenerateVlanCountExceedsMax(t *testing.T) {
	cmd := newTestRootCmd()
	_, err := executeCommand(cmd, "generate", "--vlan-count", "4094")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--vlan-count")
	assert.Contains(t, err.Error(), "4093")
}

func TestGenerateBaseConfigRejectedWithCSV(t *testing.T) {
	tmpDir := t.TempDir()
	basePath := filepath.Join(tmpDir, "base.xml")
	require.NoError(t, os.WriteFile(basePath, []byte("<?xml version=\"1.0\"?><opnsense/>"), 0o600))

	cmd := newTestRootCmd()
	_, err := executeCommand(cmd, "generate",
		"--format", "csv",
		"--base-config", basePath,
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--base-config is only supported with --format xml")
}

func TestGenerateBaseConfigMalformed(t *testing.T) {
	tmpDir := t.TempDir()
	badPath := filepath.Join(tmpDir, "bad.xml")
	require.NoError(t, os.WriteFile(badPath, []byte("not <xml properly"), 0o600))

	cmd := newTestRootCmd()
	_, err := executeCommand(cmd, "generate", "--base-config", badPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "load base config")
}

func TestGenerateOverlayReplacesFilterWholesale(t *testing.T) {
	// When --base-config is supplied and --firewall-rules is NOT set, the
	// overlay replaces the base's <filter> section with an empty one from
	// the device. Pin this behavior so a future shift to merge-semantics
	// is a deliberate, test-visible change.
	tmpDir := t.TempDir()
	basePath := filepath.Join(tmpDir, "base.xml")
	base := `<?xml version="1.0"?>
<opnsense>
  <version>1.0</version>
  <system><hostname>base</hostname><domain>base.test</domain></system>
  <filter>
    <rule>
      <type>block</type>
      <descr>from base — must be dropped on wholesale overlay</descr>
    </rule>
  </filter>
</opnsense>`
	require.NoError(t, os.WriteFile(basePath, []byte(base), 0o600))

	outPath := filepath.Join(tmpDir, "overlay.xml")
	cmd := newTestRootCmd()
	_, err := executeCommand(cmd, "generate",
		"--base-config", basePath,
		"--seed", "1",
		"--output", outPath,
	)
	require.NoError(t, err)

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	content := string(data)
	assert.NotContains(t, content, "must be dropped on wholesale overlay",
		"base Filter.Rule must be replaced wholesale when overlaying without --firewall-rules")
}

func TestGenerateHostnameAndDomainOverride(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "named.xml")

	cmd := newTestRootCmd()
	_, err := executeCommand(cmd, "generate",
		"--seed", "42",
		"--hostname", "mygateway",
		"--domain", "example.test",
		"--output", outPath,
	)
	require.NoError(t, err)

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "<hostname>mygateway</hostname>")
	assert.Contains(t, content, "<domain>example.test</domain>")
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
