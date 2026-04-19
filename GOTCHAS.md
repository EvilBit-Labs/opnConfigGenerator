# Development Gotchas & Pitfalls

This document tracks non-obvious behaviors, common pitfalls, and architectural "gotchas" in the opnConfigGenerator codebase to assist future maintainers and contributors.

It is also worthwhile to check the [opnDossier GOTCHAS](https://raw.githubusercontent.com/EvilBit-Labs/opnDossier/refs/heads/main/GOTCHAS.md) for related information.

## 1. Testing & Concurrency

## 2. Plugin Architecture

## 3. Data Processing

## 4. Generator Engine

## 5. CLI Flag Wiring

## 6. Validator

## 7. CommonDevice to Device Serializer

### 7.1 Map-backed XML sections emit in randomized order

`opnsense.Interfaces` and `opnsense.Dhcpd` in `github.com/EvilBit-Labs/opnDossier/pkg/schema/opnsense` are defined as `map[string]T` with a custom `MarshalXML` that iterates the map directly. Go map iteration is randomized per encode, so a naive `xml.Marshal(doc)` emits `<interfaces>` and `<dhcpd>` children in a different order on every call — even with a fixed RNG seed.

- **Symptom:** `generate --seed 42` produces byte-different output across runs; `TestGenerateDeterministicSeed` flakes; downstream diff tooling registers spurious changes.
- **Fix:** `internal/opnsensegen/template.go` `MarshalConfig` post-processes the marshaled XML via `sortMapBackedSections`, which walks the token stream and re-emits the children of any element in `mapBackedSections` in alphabetical order by tag name.
- **When adding a new subsystem:** if the opnDossier schema type is `map[string]T`, add the parent element name to `mapBackedSections` in `internal/opnsensegen/template.go`. Slice-backed sections (VLANs, Filter.Rule, CAs, Certs) are emitted in struct-field order and do not need this treatment.

### 7.2 Serializer must propagate every round-trip field or fail CI silently

opnDossier's `opnsenseparser.ConvertDocument` does **not** warn on structurally-valid but semantically-empty output. If `SerializeInterfaces` drops `Interface.Type` or `Interface.Virtual`, the parser reads them back as the zero value and produces zero `ConversionWarning`s. Round-trip assertions based on counts alone will pass silently while every generated VLAN interface loses its `Virtual: true` flag.

- **Fix:** `TestRoundTrip` in `internal/serializer/opnsense/serializer_test.go` asserts per-field parity on `Interface.Type`, `Virtual`, `Description`, `IPAddress`, `Subnet`; on `VLAN.Tag`, `VLANIf`, `PhysicalIf`, `Description`; on `DHCPScope.Range`, `Gateway`, `DNSServer`. Any new subsystem must extend this test or CI will not gate its round-trip fidelity.

## 8. Git Tagging

### 8.1 Tag the Squash-Merge Commit on Main

When tagging a release after a squash-merge PR, always tag the resulting commit **on `main`**, not the PR branch head. Squash-merge creates a new commit on `main` that is not an ancestor of the branch commits. If you tag the branch head instead, the tag points to an orphaned commit that `git log main` and `git describe` will never reach.

- **Symptom:** `git tag --merged main` does not list the release tag; `git describe` on `main` skips the version.
- **Fix:** `git checkout main && git pull && git tag vX.Y.Z && git push origin vX.Y.Z`
- **Prevention:** Always switch to `main` and pull before tagging. Never tag from the feature branch after merge.
