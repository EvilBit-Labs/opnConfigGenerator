// Tests in this file pin NATS-146 AC #3: consuming opnDossier's public pkg/
// API surface must not pull in CLI-only transitive dependencies. A regression
// here means opnDossier moved a CLI symbol into pkg/ or a public pkg/ now
// depends on something that was previously CLI-only.
//
// See the maintainer note on cliOnlyPackages below for how to update the
// exclusion list when opnDossier's dependency surface evolves.
package opnsensegen_test

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// cliOnlyPackages names import-path prefixes that opnDossier's CLI uses for
// display, rendering, or interactive UI and that MUST NOT be reachable from a
// consumer that only imports opnDossier's pkg/ surface.
//
// To update this list: inspect opnDossier's root go.mod for direct deps used
// exclusively by the cobra-based CLI (markdown rendering, TUI, pager, syntax
// highlighting) and add a prefix here with a comment naming the CLI surface
// that pulls it in. Keep entries conservative — only list packages with zero
// legitimate pkg/ use.
var cliOnlyPackages = []string{
	"github.com/charmbracelet/glamour",   // markdown rendering for `opnDossier render`
	"github.com/alecthomas/chroma",       // syntax highlighting, transitive of glamour
	"github.com/charmbracelet/bubbletea", // TUI framework for interactive CLI screens
	"github.com/charmbracelet/bubbles",   // TUI widgets built on bubbletea
	"github.com/olekukonko/tablewriter",  // CLI table rendering
	"github.com/muesli/reflow",           // terminal text wrapping for CLI output
}

// TestConsumerDependencyIsolation verifies that the transitive import graph of
// internal/opnsensegen — which consumes opnDossier/pkg/model and
// opnDossier/pkg/parser/opnsense — does not include any CLI-only opnDossier
// dependencies.
func TestConsumerDependencyIsolation(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping dependency graph check in -short mode")
	}

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "list", "-deps",
		"-f", "{{.ImportPath}}",
		"github.com/EvilBit-Labs/opnConfigGenerator/internal/opnsensegen")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Skipf("cannot run `go list -deps` (infrastructure issue, not a regression): %v\n%s", err, output)
	}

	deps := strings.Split(strings.TrimSpace(string(output)), "\n")
	require.NotEmpty(t, deps, "go list -deps returned no packages")

	leaked := findLeakedPackages(deps, cliOnlyPackages)
	require.Empty(t, leaked,
		"opnDossier CLI-only packages leaked into consumer transitive deps: %v\n"+
			"Run `go mod why -m <module>` on each to find the shortest path. "+
			"A leak means opnDossier moved a CLI symbol into pkg/ or a public pkg/ "+
			"gained a CLI-only transitive dependency.",
		leaked)
}

func findLeakedPackages(deps, cliOnlyPrefixes []string) []string {
	var leaked []string
	for _, dep := range deps {
		for _, prefix := range cliOnlyPrefixes {
			if strings.HasPrefix(dep, prefix) {
				leaked = append(leaked, dep)
				break
			}
		}
	}
	return leaked
}
