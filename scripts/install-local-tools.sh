#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
go_root="/Users/cosborn/.gvm/gos/go1.22.8"
go_bin="${go_root}/bin/go"

export GOROOT="${go_root}"
export GOTOOLCHAIN="go1.22.8"
export PATH="${go_root}/bin:${PATH}"
export GOBIN="${repo_root}/bin"
export GOMODCACHE="${repo_root}/.cache/gomod"

mkdir -p "${GOBIN}" "${GOMODCACHE}"

golangci_version="v1.63.4"

"${go_bin}" install "github.com/golangci/golangci-lint/cmd/golangci-lint@${golangci_version}"

echo "Installed golangci-lint ${golangci_version} to ${GOBIN}"
