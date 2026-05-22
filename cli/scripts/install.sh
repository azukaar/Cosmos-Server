#!/usr/bin/env bash
# cosmos CLI installer
# Usage: curl -fsSL https://raw.githubusercontent.com/VHStark/Cosmos-Server/main/cli/scripts/install.sh | bash

set -euo pipefail

REPO="VHStark/Cosmos-Server"
BIN_DIR="${COSMOS_CLI_BIN:-$HOME/.local/bin}"
TMP=$(mktemp -d)
trap "rm -rf $TMP" EXIT

RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
NC='\033[0m'

info()    { echo -e "${CYAN}[cosmos]${NC} $*"; }
success() { echo -e "${GREEN}[cosmos]${NC} $*"; }
die()     { echo -e "${RED}[cosmos]${NC} $*" >&2; exit 1; }

# ── detect OS / arch ──────────────────────────────────────────────────────────
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case $ARCH in
	x86_64)  ARCH="amd64" ;;
	aarch64|arm64) ARCH="arm64" ;;
	*) die "Unsupported architecture: $ARCH" ;;
esac
case $OS in
	linux|darwin) ;;
	*) die "Unsupported OS: $OS. On Windows, download the .exe from GitHub Releases." ;;
esac

# ── fetch latest release ──────────────────────────────────────────────────────
info "Fetching latest release..."
LATEST=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | grep 'cli/' | sed 's/.*"cli\/\(.*\)".*/\1/')
if [ -z "$LATEST" ]; then
	die "Could not determine latest release. Check https://github.com/${REPO}/releases"
fi

BINARY="cosmos-${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/download/cli/${LATEST}/${BINARY}"

info "Downloading cosmos CLI ${LATEST} (${OS}/${ARCH})..."
curl -fsSL -o "${TMP}/cosmos" "${URL}"
chmod +x "${TMP}/cosmos"

mkdir -p "${BIN_DIR}"
mv "${TMP}/cosmos" "${BIN_DIR}/cosmos"

success "cosmos CLI ${LATEST} installed to ${BIN_DIR}/cosmos"
echo ""

if ! echo "$PATH" | grep -q "${BIN_DIR}"; then
	echo "  Add this to your shell profile:"
	echo "    export PATH="${BIN_DIR}:\$PATH""
	echo ""
fi

echo "Quick start:"
echo "  cosmos configure"
echo "  cosmos system status"
