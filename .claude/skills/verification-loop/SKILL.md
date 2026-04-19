---
name: verification-loop
description: Verification loop tailored to opnConfigGenerator — wraps `just ci-check`, `mise exec`, and this repo's gate policies.
origin: ECC
customized: 2026-04-18 (opnConfigGenerator)
---

# Verification Loop — opnConfigGenerator

This is a Go CLI project. All verification runs through `just` recipes, and all Go commands go through `mise exec --` so the mise-pinned toolchain is used — not whatever Go happens to be on PATH.

## When to Use

- After completing a feature or significant code change
- Before creating a PR (see AGENTS.md: "CRITICAL: Run `just ci-check` BEFORE committing")
- After refactoring or dependency changes (`go.mod`/`go.sum`)
- Before release work (`just release-snapshot`, etc.)

## Primary Command

```bash
mise exec -- just ci-check
```

Expands to: `check format-check lint test test-integration test-race`. AGENTS.md §Code Quality Policy is zero-tolerance — if any step fails, investigate and fix it, including anything labelled "pre-existing."

## Granular Recipes (when triaging failures)

| Recipe                               | Purpose                                                        |
| ------------------------------------ | -------------------------------------------------------------- |
| `mise exec -- just check`            | `pre-commit run --all-files` (pre-commit gate)                 |
| `mise exec -- just format-check`     | gofumpt / goimports check                                      |
| `mise exec -- just lint`             | golangci-lint (see `.golangci.yml`)                            |
| `mise exec -- just test`             | `go test ./...`                                                |
| `mise exec -- just test-race`        | `go test -race -timeout 10m ./...` (timeout owned by justfile) |
| `mise exec -- just test-integration` | `go test -tags=integration ./...`                              |
| `mise exec -- just test-coverage`    | writes `coverage.txt`                                          |
| `mise exec -- just security-all`     | gosec + govulncheck + secret scan                              |

## Phase-by-Phase (when `ci-check` fails)

1. **Format** — `gofumpt` + `goimports`. Note: `//nolint:` directives must be on a SEPARATE LINE above the call — inline `//nolint:` gets stripped by gofumpt. See AGENTS.md §Mandatory Practices #6.
2. **Lint** — respect `.golangci.yml`. Never suppress a finding without a specific `//nolint:<linter>` and a justification comment. `testifylint` enforces `require` for error assertions and `noctx` enforces `exec.CommandContext` over `exec.Command`.
3. **Test** — all tests run with `-race` in CI via `test-race`. New tests use table-driven patterns — see `internal/opnsensegen/*_test.go` for canonical examples.
4. **Integration** — `test-integration` runs suites tagged `integration`.

## Security Scan

`mise exec -- just security-all` runs gosec, govulncheck, and a secret scan. Gosec exclusions live in `justfile` (the `-exclude=G...` flags on the gosec recipe) and in `.golangci.yml` under `linters.settings.gosec`. Do not add an exclusion without a reference to `SECURITY.md` or a documented rationale.

## Diff Review

```bash
git diff --stat
git diff <base>...HEAD
```

Review each changed file for:

- Unintended edits (especially to `mise.toml`/`mise.lock`, `go.mod`/`go.sum`, `.golangci.yml`)
- Missing error wrapping — use `fmt.Errorf("context: %w", err)`
- Secret leaks — particularly `PrivateKey`/`Password`/`PSK` fields when exporting `CommonDevice`. See `docs/solutions/best-practices/opndossier-pkg-consumer-integration-2026-04-18.md` for the redaction footgun.

## Output Format

```
VERIFICATION REPORT (opnConfigGenerator)
========================================
just ci-check: [PASS/FAIL]
  - format-check:     [PASS/FAIL]
  - lint:             [PASS/FAIL] (N findings)
  - test:             [PASS/FAIL] (X/Y passed)
  - test-integration: [PASS/FAIL]
  - test-race:        [PASS/FAIL]
Security (if run):    [PASS/FAIL]

Overall: [READY/NOT READY] for PR
Blockers: ...
```

## See Also

- `AGENTS.md` §Mandatory Practices — the canonical project contract
- `GOTCHAS.md` — non-obvious behaviors and hard-won lessons
- `CONTRIBUTING.md` — PR process and commit conventions
- `justfile` — every verification recipe lives here
