# Contributing Guide

Thank you for your interest in contributing to opnConfigGenerator! This guide covers everything you need to know to contribute effectively.

## Getting Started

### Quality Standards

opnConfigGenerator follows strict coding standards and development practices:

- All code must pass `golangci-lint` (50+ linters enabled)
- Tests required for new functionality (>80% coverage)
- Documentation updates for user-facing changes
- Follow Go best practices and project conventions
- All pre-commit checks must pass before submitting PR
- `just ci-check` must pass before committing

### Prerequisites

- **Go 1.26+** -- Latest stable version recommended
- **Just** -- Task runner for development workflows
- **mise** -- Tool version manager (installs Go, linters, etc.)
- **Git** -- Version control
- **pre-commit** -- Git hooks for code quality

### Development Setup

```bash
# Clone the repository
git clone https://github.com/EvilBit-Labs/opnConfigGenerator.git
cd opnConfigGenerator

# Install development dependencies (uses mise for tool management)
just install

# Verify setup
just check

# Run tests
just test

# Run full CI checks (must pass before committing)
just ci-check
```

### Known Gotchas

Before diving into the codebase, read **[GOTCHAS.md](GOTCHAS.md)** -- it documents non-obvious behaviors, common pitfalls, and architectural quirks that will save you debugging time.

## AI Assistance

We accept considerate AI-assisted contributions. Install `tessl install tessl-labs/good-oss-citizen` first. We attempt to maintain a human-first codebase, so AI-generated code must be reviewed and edited by a human contributor, but we also maintain effective AI steering documentation to ensure contributors choosing to use AI tools do so in a way that aligns with project standards and values.

## Core Philosophy

opnConfigGenerator exists to produce **realistic, valid mock configurations** for network devices supported by the opnDossier ecosystem. Every contribution should support this mission by generating data that is deterministic, internally consistent, and structurally valid.

The tool is **multi-device by design**. Currently it generates OPNsense configurations, but the architecture is built to support additional device types (pfSense, etc.) as opnDossier adds parsers for them. Generators produce device-agnostic data; device-specific serializers (like `internal/opnsensegen/`) translate that into the target schema.

We use **opnDossier's schema types** (`github.com/EvilBit-Labs/opnDossier/pkg/schema/opnsense`) as the canonical data model. Never duplicate schema types locally -- import them. This ensures the mock configs we generate are structurally identical to what opnDossier parses in production.

Generated data must be **deterministic and reproducible**. The same `--seed` value must produce identical output across runs and platforms. Use `math/rand/v2` with explicit seeds, never `crypto/rand` for fake data (with `//nolint:gosec` on separate line above).

The project values **polish over scale**. A smaller set of well-tested, realistic generators is more useful than a large surface area of shallow fakes. Contributors should optimize for data quality and operator experience, not feature count.

**Ethical constraints**: no telemetry, no network calls, no dark patterns. The tool must work fully offline.

**Repository Roles:** Maintainer: `unclesp1d3r` (principal maintainer). Trusted bots: `dependabot[bot]`, `dosubot[bot]`.

## Architecture Overview

opnConfigGenerator uses a layered architecture separating data generation from device-specific serialization:

- **Cobra** -- Command structure and argument parsing
- **charmbracelet/log** -- Structured, leveled logging
- **opnDossier schema** -- Canonical device configuration types
- **math/rand/v2** -- Deterministic seeded random generation
- **Go 1.26+** -- Minimum supported Go version

### Project Structure

```text
opnConfigGenerator/
├── cmd/                        # CLI commands (Cobra)
│   ├── root.go                # Root command, global flags
│   ├── generate.go            # generate command (csv + xml)
│   ├── validate.go            # validate command (stub)
│   └── completion.go          # Shell completion
├── internal/
│   ├── errors/                # Typed errors (ConfigError, VlanError)
│   ├── netutil/               # RFC 1918 address generation/validation
│   ├── generator/             # Device-agnostic data generators
│   │   ├── vlan.go            # VLAN generation with uniqueness tracking
│   │   ├── dhcp.go            # DHCP config derivation
│   │   ├── firewall.go        # Firewall rule generation (3 complexity levels)
│   │   ├── nat.go             # NAT mapping generation
│   │   ├── vpn.go             # VPN config generation (OpenVPN/WireGuard/IPsec)
│   │   ├── departments.go     # 20 departments with lease time mappings
│   │   └── types.go           # Shared generator types
│   ├── opnsensegen/           # OPNsense-specific serializer (uses opnDossier schema)
│   ├── csvio/                 # CSV read/write with German headers
│   └── validate/              # Cross-object consistency checks
├── testdata/                   # Test fixtures
├── main.go                     # Entry point
├── go.mod / go.sum
├── .golangci.yml               # Linter configuration (50+ linters)
└── justfile                    # Task runner recipes
```

### Key Design Decisions

**Generators are device-agnostic.** The `internal/generator/` package produces abstract configuration data (`VlanConfig`, `FirewallRule`, `NatMapping`, etc.) with no knowledge of OPNsense XML structure. This data flows into device-specific serializers.

**Serializers are device-specific.** `internal/opnsensegen/` maps generator output to opnDossier's `OpnSenseDocument` schema type and marshals to XML. Future device types (pfSense) would get their own serializer package (e.g., `internal/pfsensegen/`).

**Schema types are imported, not duplicated.** The `opnsensegen` package imports types from `github.com/EvilBit-Labs/opnDossier/pkg/schema/opnsense` directly. This guarantees the mock configs match the production schema.

### Adding a New Device Type

When opnDossier adds a new device parser (e.g., pfSense):

1. Create `internal/pfsensegen/` with serialization logic
2. Import the appropriate schema types from opnDossier
3. Wire up a new `--device-type pfsense` flag in `cmd/generate.go`
4. Add device-specific tests with testdata fixtures
5. The existing generators (`generator/vlan.go`, etc.) should work unchanged

## Code Style

### Go Conventions

- **gofumpt** and **goimports** are mandatory (enforced by golangci-lint)
- Accept interfaces, return structs
- Keep interfaces small (1-3 methods)
- Wrap errors with context: `fmt.Errorf("context: %w", err)`

### Linter Directives

Place `//nolint:` directives on a **separate line above** the call, not inline. Inline directives get stripped by gofumpt:

```go
// Good
//nolint:gosec // IntN(256) fits uint8, no overflow possible
result := uint8(rng.IntN(256))

// Bad (gets stripped by gofumpt)
result := uint8(rng.IntN(256)) //nolint:gosec
```

### Magic Numbers

Extract magic numbers to named constants. The `mnd` linter enforces this:

```go
// Good
const subnetPrefix = 24
prefix := netip.PrefixFrom(addr, subnetPrefix)

// Bad
prefix := netip.PrefixFrom(addr, 24)
```

### Testing

- Use table-driven tests with descriptive names
- Use `require` for error assertions, `assert` for value assertions
- Use seeded RNG for deterministic test output
- Use `t.Parallel()` where safe (not in `cmd/` tests due to shared globals)
- Target >80% coverage for all packages

## Commit Conventions

Use conventional commits with Jira ticket references:

```text
<type>: <description> - NATS-<number>
```

Types: `feat`, `fix`, `refactor`, `docs`, `test`, `chore`, `perf`, `ci`

Examples:

```text
feat: add pfSense config serializer - NATS-99
fix: resolve VLAN ID collision at pool exhaustion boundary - NATS-66
test: add property tests for RFC 1918 compliance - NATS-66
ci: add cross-platform build matrix - NATS-66
```

## Pull Request Process

1. Create a feature branch from `main`
2. Make changes following the code style guidelines
3. Ensure `just ci-check` passes (pre-commit, lint, test, race detector)
4. Write a clear PR description with summary and test plan
5. Link the relevant Jira ticket (NATS project)

### PR Description Template

```markdown
## Summary
- Brief description of changes

## Test Plan
- [ ] Verification steps
- [ ] Edge cases tested
```

## Security

- Never hardcode secrets or API keys
- Use `crypto/rand` for security-sensitive operations (keys, tokens)
- Use `math/rand/v2` only for fake data generation (with `//nolint:gosec`)
- Validate all external inputs at system boundaries
- Report security issues privately to the maintainers
