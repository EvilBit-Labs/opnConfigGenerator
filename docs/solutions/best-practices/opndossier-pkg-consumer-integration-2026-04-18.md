---
title: Consuming opnDossier's public pkg/ API as an external Go module
date: 2026-04-18
category: best-practices
module: internal/opnsensegen
problem_type: best_practice
component: tooling
severity: critical
related_components:
  - testing_framework
  - documentation
applies_when:
  - Importing opnDossier's pkg/model, pkg/parser, or pkg/schema types from an external Go module
  - Converting a generated OPNsense config.xml into *model.CommonDevice without re-implementing the parser
  - Avoiding transitive CLI dependencies (glamour, chroma, bubbletea, bubbles, tablewriter, reflow) in a library/CLI consumer
  - Choosing between pkg/parser.Factory.CreateDevice and pkg/parser/opnsense.ConvertDocument as the ingestion entrypoint
  - Adding a regression guard that walks go list -deps to prove dependency isolation
  - Exporting or serializing CommonDevice from a consumer (JSON, YAML, XML) without leaking credentials
tags:
  - opndossier
  - pkg-consumer
  - commondevice
  - go-modules
  - dependency-isolation
  - xml-parsing
  - privatekey-redaction
  - nats-146
---

# Consuming opnDossier's public pkg/ API as an external Go module

## Context

opnDossier's `pkg/` tree is a public Go API surface that external consumers — including opnConfigGenerator — are expected to import. The surface includes `pkg/model` (the platform-agnostic `CommonDevice` and related types), `pkg/parser` (the `Factory` + `XMLDecoder` interface + parser registry), `pkg/parser/opnsense` and `pkg/parser/pfsense` (device-specific parsers with `ConvertDocument` convenience functions), and `pkg/schema/{opnsense,pfsense}` (raw XML schema types). Before NATS-146, no external consumer had exercised the full file → `CommonDevice` pipeline from outside the opnDossier repo, so the consumer contract was asserted but not verified.

The boundary is deliberate: *"Anything that operates on CommonDevice should stay in opnDossier. The public surface covers file → CommonDevice only."* That means `pkg/` holds parsing and conversion; `internal/sanitizer`, `internal/converter` (enrichment), audit plugins, diff engine, and report builders stay private. Consumers who need those must stay inside opnDossier.

## Guidance

### 1. Pick the right ingestion entrypoint

Both paths are valid public APIs — choose based on what you already have:

| Entrypoint                                                                  | Use when                                                                                                       | Requires                                                                                                                                                                                                        |
| --------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `pkg/parser.Factory.CreateDevice(ctx, r, deviceTypeOverride, validateMode)` | You have XML bytes or a `Reader` and want auto-detection of the device type                                    | Consumer-supplied `parser.XMLDecoder` implementation (opnDossier keeps its concrete decoder in `internal/cfgparser`); blank import of `pkg/parser/opnsense` and/or `pkg/parser/pfsense` for parser registration |
| `pkg/parser/opnsense.ConvertDocument(doc *schema.OpnSenseDocument)`         | You already have a parsed `*schema/opnsense.OpnSenseDocument` (via `encoding/xml` or your own schema pipeline) | Nothing — direct function call                                                                                                                                                                                  |

opnConfigGenerator chose `ConvertDocument` because it already parses XML through its own `opnsensegen.ParseConfig` (thin `encoding/xml` wrapper). Writing an `XMLDecoder` adapter for `Factory` would add code without new coverage.

```go
import (
    "github.com/EvilBit-Labs/opnDossier/pkg/model"
    opnsenseparser "github.com/EvilBit-Labs/opnDossier/pkg/parser/opnsense"
)

device, warnings, err := opnsenseparser.ConvertDocument(parsed)
if err != nil {
    return fmt.Errorf("convert to CommonDevice: %w", err)
}
```

### 2. If using `Factory`, respect the registration contract

`pkg/parser` follows a `database/sql`-style driver registration pattern. Device parsers register themselves via `init()` inside `pkg/parser/opnsense` and `pkg/parser/pfsense`. You must blank-import at least one of them or `Factory.CreateDevice` returns an "empty registry" error. opnDossier's registry currently returns an error that names the blank import to add; if that contract changes, `Factory` users will see a generic empty-registry error instead — treat the hint as best-effort, not guaranteed.

### 3. Round-trip verification tests through XML bytes, not in-memory structs

A consumer-facing test that marshals to XML and re-parses proves the pipeline works against on-disk representation (which is what real consumers do) and catches encoder/decoder asymmetry. In-memory-only tests miss marshal round-trip bugs.

### 4. Pin dependency isolation as a programmatic regression test

`go mod why` is a manual diagnostic. An in-repo test that shells out to `go list -deps` and asserts no CLI-only modules appear in the transitive graph of your consumer package enforces the contract every CI run. Keep the CLI-only exclusion list maintainer-annotated — hardcode the prefixes with comments naming the CLI surface that pulls each one in, and point updaters at opnDossier's root `go.mod` as the source of truth.

### 5. Pin the nil-document error contract

`ConvertDocument(nil)` returns the exported sentinel `opnsenseparser.ErrNilDocument`. Assert against the sentinel with `require.ErrorIs` so a silent contract change (different error, panic) fails loudly in CI.

### 6. CRITICAL: redact before exporting a CommonDevice

**`pkg/model/CommonDevice` contains credential fields that opnDossier's public surface does NOT redact automatically.** A consumer who marshals `CommonDevice` directly to JSON, YAML, or XML **will leak secrets.**

The redaction logic lives in `internal/sanitizer/` and `internal/converter/` on the opnDossier side and is not exported via `pkg/`. Public API consumers must implement their own redaction pass before writing a `CommonDevice` to disk, logs, or a network response.

Known secret-bearing fields in `pkg/model` as of opnDossier v1.4.0 (non-exhaustive — re-audit on every opnDossier bump):

| Field                             | File                        | Notes                    |
| --------------------------------- | --------------------------- | ------------------------ |
| `Certificate.PrivateKey`          | `pkg/model/certificates.go` | TLS private key material |
| `CertificateAuthority.PrivateKey` | `pkg/model/certificates.go` | CA private key           |
| `WireGuardClient.PSK`             | `pkg/model/vpn.go`          | Pre-shared key           |
| `APIKey.Secret`                   | `pkg/model/users.go`        | API secret               |
| `HighAvailability.Password`       | `pkg/model/ha.go`           | XMLRPC sync password     |
| `Bindpw`, `ROCommunity`, etc.     | various                     | LDAP/SNMP credentials    |

Until opnDossier exports public redaction helpers on `CommonDevice`, treat the unredacted struct as secret-bearing. Check the upstream `pkg/model/` source for any new `*Key`, `*Password`, `*Secret`, `PSK`, `Community` fields introduced since v1.4.0 before releasing a consumer who serializes the struct.

### 7. Handle `ConversionWarning`s explicitly

`pkg/model/warning.go` defines `ConversionWarning` (with a `Severity` enum, NATS-145) as the public warning type. `ConvertDocument` returns `[]ConversionWarning` alongside the device. For well-formed input that the consumer controls, expect zero warnings — treat non-empty as a diagnostic signal and surface the warnings in the failure message rather than discarding them, so regressions in generator output are debuggable.

### 8. Watch for stack-specific quirks

- **pfSense parser ignores the injected `XMLDecoder`** — pfSense's `Parser` self-manages XML decoding. Consumers using the `Factory` path should know this; the injected decoder is respected by the OPNsense parser only.
- **`KeaDhcp4` decodes silently** — if it matters to your consumer, check for empty fields after conversion.

## Why This Matters

Without a documented entrypoint, new consumers reach for `pkg/parser.Factory` first, hit the missing-decoder wall (opnDossier keeps its decoder in `internal/cfgparser`), and either reimplement XML parsing or give up and import the internal package via a fork — both bad outcomes. Without the dependency-isolation test, a future opnDossier refactor that moves a CLI helper into `pkg/` silently inflates consumer binaries by tens of megabytes of TUI/rendering dependencies; the regression is caught at `go build` time only after it ships. Without the redaction warning, the first external consumer to export a `CommonDevice` to a file or API surface leaks private keys on day one.

Treating the public surface as a contract with programmatic tests — not prose in a README — means opnDossier can evolve its internal layout without breaking external consumers, and consumers get early warning when the contract drifts.

## When to Apply

- Adding a new Go package that imports `github.com/EvilBit-Labs/opnDossier/pkg/...`
- Building a feature that needs a normalized `*model.CommonDevice` from an OPNsense config.xml
- Upgrading the opnDossier dependency across a minor or major version — re-run the isolation test and scan `pkg/model/` for new secret-bearing fields
- Writing an external tool, service, or test harness that consumes opnDossier as a library rather than a CLI
- Reviewing a PR that touches `go.mod` or adds an `opnDossier/pkg/...` import
- Before exporting, logging, or persisting a `CommonDevice` from a consumer — audit for unredacted credentials
- Onboarding a contributor who asks "how do I turn config.xml into CommonDevice from outside opnDossier?"

## Examples

### Minimal `ConvertDocument` call

From `internal/opnsensegen/commondevice_test.go`:

```go
cfg, err := opnsensegen.LoadBaseConfig("../../testdata/base-config.xml")
require.NoError(t, err)

device, _, err := opnsenseparser.ConvertDocument(cfg)
require.NoError(t, err)
assert.Equal(t, model.DeviceTypeOPNsense, device.DeviceType)
```

### Full XML round-trip (the consumer acceptance shape)

```go
// Build/mutate the document in memory.
cfg, _ := opnsensegen.LoadBaseConfig("../../testdata/base-config.xml")
cfg.System.Hostname = "nats146-host"
opnsensegen.InjectVlans(cfg, vlans, 6)

// Round-trip through XML bytes — proves the pipeline works against
// on-disk representation, not just in-memory struct passing.
var buf bytes.Buffer
require.NoError(t, opnsensegen.MarshalConfig(cfg, &buf))

parsed, err := opnsensegen.ParseConfig(buf.Bytes())
require.NoError(t, err)

device, warnings, err := opnsenseparser.ConvertDocument(parsed)
require.NoError(t, err)
assert.Empty(t, warnings,
    "generator output produced %d ConversionWarning(s): %+v", len(warnings), warnings)
```

### Dependency-isolation regression test

From `internal/opnsensegen/deps_isolation_test.go`. Two details matter:

1. **`go list -deps -test`** — the opnDossier consumer imports live in `_test.go` files, so a plain `go list -deps <pkg>` excludes them. Without `-test`, the check silently passes even if `pkg/model` or `pkg/parser/opnsense` gained a CLI-only transitive dep — defeating the whole point.
2. **Gate under `//go:build integration`** — the test shells out to the Go toolchain, so it belongs in `just test-integration` (which `just ci-check` runs) rather than the hot path of `just test`.

```go
//go:build integration

var cliOnlyPackages = []string{
    "github.com/charmbracelet/glamour",   // markdown rendering for `opnDossier render`
    "github.com/alecthomas/chroma",       // syntax highlighting, transitive of glamour
    "github.com/charmbracelet/bubbletea", // TUI framework
    "github.com/charmbracelet/bubbles",   // TUI widgets
    "github.com/olekukonko/tablewriter",  // CLI table rendering
    "github.com/muesli/reflow",           // terminal text wrapping
    // To update: inspect opnDossier's root go.mod for direct deps used
    // exclusively by the cobra CLI. Keep entries conservative.
}

func TestConsumerDependencyIsolation(t *testing.T) {
    t.Parallel()

    ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
    defer cancel()

    cmd := exec.CommandContext(ctx, "go", "list", "-deps", "-test",
        "-f", "{{.ImportPath}}",
        "github.com/EvilBit-Labs/opnConfigGenerator/internal/opnsensegen")
    output, err := cmd.CombinedOutput()
    if err != nil {
        // Only SKIP when the go toolchain is missing. Anything else
        // (non-zero exit, context timeout, broken go.mod) is a real
        // regression signal and must FAIL the test.
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

    // Sanity check: without this, a broken -test flag would yield a
    // false pass because pkg/model/pkg/parser wouldn't be in the graph.
    for _, want := range []string{
        "github.com/EvilBit-Labs/opnDossier/pkg/model",
        "github.com/EvilBit-Labs/opnDossier/pkg/parser/opnsense",
    } {
        require.Contains(t, deps, want, "expected %q in transitive deps", want)
    }

    leaked := findLeakedPackages(deps, cliOnlyPackages)
    require.Empty(t, leaked,
        "opnDossier CLI-only packages leaked into consumer transitive deps: %v\n"+
            "Run `go mod why -m <module>` to find the shortest path.", leaked)
}
```

### Nil-input contract pin

```go
device, warnings, err := opnsenseparser.ConvertDocument(nil)
require.ErrorIs(t, err, opnsenseparser.ErrNilDocument)
assert.Nil(t, device)
assert.Empty(t, warnings)
```

### Redaction wrapper sketch (consumer responsibility)

Until opnDossier exports public redaction helpers, consumers that serialize `CommonDevice` should wrap it:

```go
// RedactForExport clears secret-bearing fields before marshaling.
// Keep this aligned with pkg/model/ — re-audit on every opnDossier bump.
func RedactForExport(d *model.CommonDevice) *model.CommonDevice {
    out := *d // shallow copy
    for i := range out.Certificates {
        out.Certificates[i].PrivateKey = ""
    }
    for i := range out.CertificateAuthorities {
        out.CertificateAuthorities[i].PrivateKey = ""
    }
    for i := range out.WireGuardClients {
        out.WireGuardClients[i].PSK = ""
    }
    // ... mirror for APIKey.Secret, HighAvailability.Password, Bindpw, ROCommunity, etc.
    return &out
}
```

This is a stopgap — track [opnDossier#NATS-146 follow-up](https://evilbitlabs.atlassian.net/browse/NATS-146) for public redaction helpers on the opnDossier side.

## Related

- Tickets: [NATS-146](https://evilbitlabs.atlassian.net/browse/NATS-146) (this work), [NATS-107 epic](https://evilbitlabs.atlassian.net/browse/NATS-107), NATS-3 / NATS-144 / NATS-145 (upstream audit chain)
- `internal/opnsensegen/commondevice_test.go` — consumer pipeline tests (round-trip, minimal, nil)
- `internal/opnsensegen/deps_isolation_test.go` — CLI-dep leak regression test
- `internal/opnsensegen/template.go` — `ParseConfig`, `MarshalConfig`, `LoadBaseConfig` helpers
- `CONTRIBUTING.md` — already forbids duplicating opnDossier schema types locally
- `GOTCHAS.md` §7 "CommonDevice to Device Serializer" — reserved for any `ConversionWarning`s or serialization quirks observed in consumer runs
- Upstream [docs/development/public-api.md @ v1.4.0](https://github.com/EvilBit-Labs/opnDossier/blob/v1.4.0/docs/development/public-api.md) — authoritative API stability policy (pin re-audit on every opnDossier bump)
