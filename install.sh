#!/usr/bin/env bash
# Builds and installs kvex from source into your Go bin directory.
#
# Usage: ./install.sh
set -euo pipefail

repo_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$repo_dir"

require_go_version="1.25.0"

info()  { printf '\033[1;34m==>\033[0m %s\n' "$1"; }
warn()  { printf '\033[1;33m!!\033[0m %s\n' "$1"; }
fail()  { printf '\033[1;31mERROR:\033[0m %s\n' "$1" >&2; exit 1; }

if ! command -v go >/dev/null 2>&1; then
  fail "Go is not installed. Install it from https://go.dev/dl/ and re-run this script."
fi

go_version="$(go env GOVERSION | sed 's/^go//')"
if [ "$(printf '%s\n%s\n' "$require_go_version" "$go_version" | sort -V | head -n1)" != "$require_go_version" ]; then
  fail "kvex requires Go >= $require_go_version, found $go_version."
fi

info "Building and installing kvex (go install ./cmd/kvex)..."
go install ./cmd/kvex

gobin="$(go env GOBIN)"
if [ -z "$gobin" ]; then
  gobin="$(go env GOPATH)/bin"
fi
kvex_path="$gobin/kvex"

if [ ! -x "$kvex_path" ]; then
  fail "Install finished but $kvex_path was not found. Check the go install output above."
fi

info "Installed: $kvex_path"

case ":$PATH:" in
  *":$gobin:"*) ;;
  *)
    warn "$gobin is not on your PATH."
    echo "  Add this to your shell profile (~/.zshrc, ~/.bashrc, ...):"
    echo "    export PATH=\"$gobin:\$PATH\""
    ;;
esac

config_dir="$HOME/.config/kvex"
config_file="$config_dir/config.yaml"
if [ ! -f "$config_file" ]; then
  mkdir -p "$config_dir"
  cp "$repo_dir/examples/config.yaml" "$config_file"
  info "Wrote a starter config to $config_file — edit it with your real vault URIs."
else
  info "Existing config found at $config_file, leaving it untouched."
fi

if ! command -v az >/dev/null 2>&1; then
  warn "Azure CLI (az) not found. kvex authenticates via DefaultAzureCredential (managed identity, env vars, etc. also work), but 'az login' is the easiest local option."
fi

info "Done. Run: kvex --config $config_file"
