# opnConfigGenerator Justfile
set shell := ["bash", "-cu"]
set windows-powershell := true
set dotenv-load := true
set ignore-comments := true

mise_exec := "mise exec --"
project_dir := justfile_directory()
binary_name := "opnconfiggenerator"
_null := if os_family() == "windows" { "nul" } else { "/dev/null" }

[private]
default:
    @just --list --unsorted

alias h := help
alias l := list

[group('help')]
help:
    @just --list

[group('help')]
list group="":
    @just --list --unsorted {{ if group != "" { "--list-heading='' --list-prefix='  ' | grep -A999 '" + group + "'" } else { "" } }}

alias i := install

[group('setup')]
install:
    @mise install
    @{{ mise_exec }} go mod tidy

[group('setup')]
setup: install

[group('setup')]
update-deps:
    @{{ mise_exec }} go get -u ./...
    @{{ mise_exec }} go mod tidy
    @{{ mise_exec }} go mod verify

alias r := run

[group('dev')]
run *args:
    @{{ mise_exec }} go run main.go {{ args }}

alias f := format
alias fmt := format

[group('quality')]
format:
    @{{ mise_exec }} golangci-lint run --fix ./...

[group('quality')]
format-check:
    @{{ mise_exec }} golangci-lint fmt ./...

[group('quality')]
lint:
    @{{ mise_exec }} golangci-lint run ./...

alias t := test

[group('test')]
test:
    @{{ mise_exec }} go test ./...

[group('test')]
test-v:
    @{{ mise_exec }} go test -v ./...

[group('test')]
test-coverage:
    @{{ mise_exec }} go test -coverprofile=coverage.txt ./...
    @{{ mise_exec }} go tool cover -func=coverage.txt

[group('test')]
test-race:
    @{{ mise_exec }} go test -race -timeout 10m ./...

[group('test')]
coverage:
    @{{ mise_exec }} go test -coverprofile=coverage.txt ./...
    @{{ mise_exec }} go tool cover -html=coverage.txt

[group('test')]
bench:
    @{{ mise_exec }} go test -bench=. -benchmem ./...

alias b := build

[group('build')]
build:
    @{{ mise_exec }} go build -o {{ binary_name }}{{ if os_family() == "windows" { ".exe" } else { "" } }} main.go

[group('build')]
build-release:
    @CGO_ENABLED=0 {{ mise_exec }} go build -trimpath -ldflags="-s -w" -o {{ binary_name }}{{ if os_family() == "windows" { ".exe" } else { "" } }} main.go

[group('build')]
[confirm("This will remove build artifacts. Continue?")]
clean:
    @{{ mise_exec }} go clean
    @rm -f coverage.txt {{ binary_name }} {{ binary_name }}.exe 2>{{ _null }} || true

[group('ci')]
ci-check: format-check lint test test-race

[group('ci')]
ci-smoke:
    @{{ mise_exec }} go build -trimpath -ldflags="-s -w -X main.version=dev" -v ./...
    @{{ mise_exec }} go test -count=1 -failfast -short -timeout 5m ./...