# opnConfigGenerator

[![Go Version][go-badge]][go] [![License][license-badge]][license]

## Overview

opnConfigGenerator is a command-line tool for generating realistic, valid network device configuration files populated with fake data. It produces OPNsense `config.xml` files suitable for testing, training, development, and demonstration purposes -- without exposing sensitive network information.

Part of the [opnDossier](https://github.com/EvilBit-Labs/opnDossier) ecosystem. Uses opnDossier's canonical schema types to ensure generated configurations are structurally identical to real device configs.

### Features

- Generate realistic VLAN configurations with RFC 1918 compliance
- Create valid interface assignments and DHCP pools
- Generate firewall rules at 3 complexity levels (basic, intermediate, advanced)
- Create VPN configurations (OpenVPN, WireGuard, IPsec)
- Generate NAT rules and port forwards
- Deterministic output with `--seed` for reproducible builds
- CSV and XML output formats
- Single static binary, no runtime dependencies, fully offline

## Quick Start

### Installation

Download pre-built binaries from [releases](https://github.com/EvilBit-Labs/opnConfigGenerator/releases), or install from source:

```bash
go install github.com/EvilBit-Labs/opnConfigGenerator@latest
```

### Usage

```bash
# Generate 25 VLANs as OPNsense XML
opnconfiggenerator generate --format xml --count 25 --base-config config.xml --seed 42

# Generate CSV data for spreadsheet analysis
opnconfiggenerator generate --format csv --count 50 --output network-data.csv

# Generate with firewall rules and VPN configs
opnconfiggenerator generate --format xml --count 15 \
  --base-config config.xml \
  --include-firewall-rules --firewall-rule-complexity advanced \
  --vpn-count 3 --nat-mappings 10 \
  --seed 42

# Validate a generated config
opnconfiggenerator validate --input generated-config.xml
```

### CLI Flags

| Flag                         | Default    | Description                                                       |
| ---------------------------- | ---------- | ----------------------------------------------------------------- |
| `--format`                   | (required) | Output format: `csv` or `xml`                                     |
| `--count`                    | 10         | Number of VLANs to generate (1-4085)                              |
| `--base-config`              |            | Base OPNsense XML template (required for xml)                     |
| `--seed`                     | 0 (random) | RNG seed for reproducible output                                  |
| `--include-firewall-rules`   | false      | Generate firewall rules per VLAN                                  |
| `--firewall-rule-complexity` | basic      | Rule complexity: `basic` (3), `intermediate` (7), `advanced` (15) |
| `--vpn-count`                | 0          | Number of VPN configurations to generate                          |
| `--nat-mappings`             | 0          | Number of NAT rules to generate                                   |
| `--wan-assignments`          | single     | WAN distribution: `single`, `multi`, `balanced`                   |
| `--output`                   | stdout     | Output file path                                                  |
| `--force`                    | false      | Overwrite existing output files                                   |
| `--quiet`                    | false      | Suppress output except errors                                     |
| `--no-color`                 | false      | Disable colored output                                            |

## Architecture

```text
Generator Layer (device-agnostic)     Serializer Layer (device-specific)
+------------------+                  +-------------------+
| generator/       |                  | opnsensegen/      |
|   vlan.go        | --- VlanConfig ---->  template.go    |---> config.xml
|   firewall.go    | --- FirewallRule ->   (uses opnDossier|
|   dhcp.go        | --- DhcpConfig --->    schema types) |
|   nat.go         | --- NatMapping -->                   |
|   vpn.go         | --- VpnConfig --->                   |
+------------------+                  +-------------------+
                                      | pfsensegen/       |  (future)
                                      +-------------------+
```

Generators produce device-agnostic data. Serializers map to device-specific schemas imported from opnDossier. Adding a new device type means adding a new serializer package -- generators stay unchanged.

## Development

```bash
# Install dependencies
just install

# Run tests
just test

# Run linter
just lint

# Run full CI checks (required before committing)
just ci-check

# Build binary
just build
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed development guidelines.

## Related Projects

- [opnDossier](https://github.com/EvilBit-Labs/opnDossier) -- OPNsense/pfSense configuration documentation and compliance auditing tool

## License

[Apache-2.0](LICENSE)

<!-- Badge links -->

[go]: https://go.dev
[go-badge]: https://img.shields.io/github/go-mod/go-version/EvilBit-Labs/opnConfigGenerator
[license]: https://github.com/EvilBit-Labs/opnConfigGenerator/blob/main/LICENSE
[license-badge]: https://img.shields.io/github/license/EvilBit-Labs/opnConfigGenerator
