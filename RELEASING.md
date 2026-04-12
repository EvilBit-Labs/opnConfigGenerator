# Releasing opnConfigGenerator

This document describes the release process for opnConfigGenerator.

## Version Numbering

opnConfigGenerator follows [Semantic Versioning 2.0.0](https://semver.org/):

| Version Component | When to Increment                                  | Example                                         |
| ----------------- | -------------------------------------------------- | ----------------------------------------------- |
| **MAJOR** (X.0.0) | Breaking changes to CLI or output format           | Removing a flag, changing CSV/XML output schema |
| **MINOR** (0.X.0) | New features, backward-compatible additions        | New device type, new generator, new output flag |
| **PATCH** (0.0.X) | Bug fixes, documentation, performance improvements | Fix VLAN collision, typo fixes                  |

### Pre-release Tags

- **Release Candidates**: `v0.2.0-rc1`, `v0.2.0-rc2` -- Feature-complete, needs testing
- **Beta**: `v0.2.0-beta.1` -- Feature incomplete, early testing

## Prerequisites

### Required Tools

Install these tools before creating a release:

```bash
# Install via mise (recommended -- see mise.toml)
mise install

# Or install manually:
brew install goreleaser/tap/goreleaser
brew install git-cliff
brew install cosign
go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest
go install github.com/google/go-licenses@latest
```

### GitHub Permissions

The release workflow requires:

- `contents: write` -- Create releases and upload assets
- `id-token: write` -- SLSA provenance and Cosign keyless signing

### GitHub Secrets

| Secret            | Description                                                                         |
| ----------------- | ----------------------------------------------------------------------------------- |
| `GPG_PRIVATE_KEY` | Base64-encoded GPG private key (`gpg --armor --export-secret-keys EMAIL \| base64`) |
| `GPG_PASSPHRASE`  | Passphrase for the GPG key                                                          |

GPG signing is optional. If these secrets are not set, releases will still be created with Cosign signatures for checksums.

## Pre-release Checklist

Before creating a release, verify:

- [ ] All CI checks pass on `main` branch
- [ ] All issues/PRs for the milestone are closed
- [ ] `just ci-check` passes locally
- [ ] Documentation reflects new features/changes
- [ ] Breaking changes are documented

### Verify CI Status

```bash
gh run list --branch main --limit 5
```

## Creating a Release

### Step 1: Validate Configuration

```bash
# Check goreleaser configuration
goreleaser check

# Preview what would be built (no publish)
goreleaser release --snapshot --clean

# Check generated artifacts
ls -la dist/
```

### Step 2: Generate Changelog Preview

```bash
# Preview changelog for unreleased commits
git-cliff --unreleased

# Preview full changelog
git-cliff --output /dev/stdout
```

### Step 3: Create and Push Tag

```bash
# Ensure you're on main with latest changes
git checkout main
git pull origin main

# Create annotated tag
git tag -a v0.1.0 -m "Release v0.1.0"

# Push tag to trigger release workflow
git push origin v0.1.0
```

### Step 4: Create GitHub Release

```bash
# Create release from tag (triggers workflow)
gh release create v0.1.0 \
  --title "v0.1.0" \
  --generate-notes
```

### Step 5: Monitor Release Workflow

```bash
gh run watch
```

## Post-release Verification

After the release workflow completes:

### Verify Artifacts

```bash
# List release assets
gh release view v0.1.0

# Download and verify checksums
gh release download v0.1.0 --pattern "*checksums*"
sha256sum -c opnConfigGenerator_checksums.txt
```

### Verify Cosign Signatures

```bash
cosign verify-blob \
  --certificate-identity "https://github.com/EvilBit-Labs/opnConfigGenerator/.github/workflows/release.yml@refs/tags/v0.1.0" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  --bundle opnConfigGenerator_checksums.txt.sigstore.json \
  opnConfigGenerator_checksums.txt
```

### Test Installation

```bash
# Test binary download and execution
gh release download v0.1.0 --pattern "*Darwin_arm64*"
tar -xzf opnConfigGenerator_Darwin_arm64.tar.gz
./opnconfiggenerator --version
```

## Hotfix Process

For urgent fixes to a released version:

```bash
# Branch from the release tag
git checkout -b hotfix/v0.1.1 v0.1.0

# Make the fix, commit, push, create PR targeting main
# After merge, tag and release
git checkout main
git pull
git tag -a v0.1.1 -m "Hotfix release v0.1.1"
git push origin v0.1.1
```

## Release Artifacts

Each release includes:

| Artifact                                         | Description                            |
| ------------------------------------------------ | -------------------------------------- |
| `opnConfigGenerator_<OS>_<arch>.tar.gz`          | Binary archives (Linux, macOS)         |
| `opnConfigGenerator_<OS>_<arch>.zip`             | Binary archive (Windows)               |
| `opnConfigGenerator_checksums.txt`               | SHA256 checksums for all artifacts     |
| `opnConfigGenerator_checksums.txt.sigstore.json` | Cosign v3 signature bundle             |
| `*.bom.json`                                     | Software Bill of Materials (CycloneDX) |

## Quick Release Checklist

- [ ] CI green on `main` -- `gh run list --branch main --limit 5`
- [ ] `just ci-check` passes locally
- [ ] Generate changelog -- `just changelog-version vX.Y.Z`
- [ ] Commit changelog to `main` and push
- [ ] Create annotated tag -- `git tag -a vX.Y.Z -m "Release vX.Y.Z"`
- [ ] Push tag -- `git push origin vX.Y.Z`
- [ ] Create release -- `gh release create vX.Y.Z --title "vX.Y.Z" --generate-notes`
- [ ] Monitor workflow -- `gh run watch`
- [ ] Verify artifacts and signatures
- [ ] Test binary -- `./opnconfiggenerator --version`
