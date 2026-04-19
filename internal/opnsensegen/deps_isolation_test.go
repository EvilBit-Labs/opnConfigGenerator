//go:build integration

// Tests in this file pin NATS-146 AC #3: consumers of opnDossier's public
// pkg/ API surface must not pull in CLI-only transitive dependencies via
// that surface. A regression here means opnDossier moved a CLI symbol into
// pkg/ or a public pkg/ now depends on something that was previously
// CLI-only.
//
// The opnDossier imports (pkg/model, pkg/parser/opnsense) live in the
// consumer test file `commondevice_test.go` — they are intentionally not
// used from production code in this package. So this check runs
// `go list -deps -test` on internal/opnsensegen so the dependency graph
// includes the test-only consumer imports we are trying to pin.
//
// This test shells out to the `go` toolchain to inspect the module graph,
// which is heavier than a normal unit test. It runs under the `integration`
// build tag (via `just test-integration`, which `just ci-check` invokes).
//
// See the maintainer note on cliOnlyPackages below for how to update the
// exclusion list when opnDossier's dependency surface evolves.
package opnsensegen_test

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// cliOnlyPackages names import-path prefixes that opnDossier's CLI uses for
// display, rendering, or interactive UI and that MUST NOT be reachable from a
// consumer who only imports opnDossier's pkg/ surface.
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

// TestConsumerDependencyIsolation verifies that the transitive import graph
// of internal/opnsensegen *including its test dependencies* does not include
// any CLI-only opnDossier dependencies. The `-test` flag is required because
// the opnDossier consumer imports (pkg/model, pkg/parser/opnsense) live in
// consumer test files, not production code.
func TestConsumerDependencyIsolation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "list", "-deps", "-test",
		"-f", "{{.ImportPath}}",
		"github.com/EvilBit-Labs/opnConfigGenerator/internal/opnsensegen")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Only skip when the go toolchain itself is missing — that is
		// an environmental issue, not a regression this test can catch.
		// Everything else (non-zero exit from go list, context timeout,
		// broken go.mod) is a real signal and must fail the test.
		var pathErr *exec.Error
		if errors.Is(err, exec.ErrNotFound) ||
			(errors.As(err, &pathErr) && errors.Is(pathErr.Err, exec.ErrNotFound)) {
			t.Skipf("go toolchain unavailable: %v", err)
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			t.Fatalf("go list -deps -test timed out after 30s: %v\n%s", err, output)
		}
		t.Fatalf("go list -deps -test failed: %v\n%s", err, output)
	}

	trimmed := strings.TrimSpace(string(output))
	require.NotEmpty(t, trimmed, "go list -deps -test returned no packages")
	deps := strings.Split(trimmed, "\n")

	// Sanity check: the opnDossier public consumer imports MUST be in the
	// graph. If they aren't, the -test flag isn't working as expected and
	// the leak check below would give a false pass.
	requiredConsumerDeps := []string{
		"github.com/EvilBit-Labs/opnDossier/pkg/model",
		"github.com/EvilBit-Labs/opnDossier/pkg/parser/opnsense",
	}
	for _, want := range requiredConsumerDeps {
		require.Contains(t, deps, want,
			"expected %q in transitive deps; `go list -deps -test` output may be wrong", want)
	}

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
		if dep == "" {
			continue
		}
		for _, prefix := range cliOnlyPrefixes {
			if dep == prefix || strings.HasPrefix(dep, prefix+"/") {
				leaked = append(leaked, dep)
				break
			}
		}
	}
	return leaked
}
