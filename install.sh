#!/usr/bin/env bash
set -euo pipefail

BIN_NAME="bqa"
INSTALL_DIR="${BQA_INSTALL_DIR:-$HOME/.local/bin}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

info() { printf "\033[1;34m%s\033[0m\n" "$*"; }
warn() { printf "\033[1;33m%s\033[0m\n" "$*"; }
fail() { printf "\033[1;31m%s\033[0m\n" "$*" >&2; exit 1; }

info "Installing BQA-OS from local checkout..."

if ! command -v go >/dev/null 2>&1; then
  fail "Go is required for this installer. Install it first: brew install go"
fi

mkdir -p "$INSTALL_DIR"

info "Building $BIN_NAME from $SCRIPT_DIR/cmd/bqa"
GOBIN="$INSTALL_DIR" go install "$SCRIPT_DIR/cmd/bqa"

if ! command -v "$INSTALL_DIR/$BIN_NAME" >/dev/null 2>&1; then
  fail "Installation failed: $INSTALL_DIR/$BIN_NAME was not created"
fi

info "Installed: $INSTALL_DIR/$BIN_NAME"

case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *)
    warn "$INSTALL_DIR is not in PATH. Add this to your shell profile:"
    echo "export PATH=\"$INSTALL_DIR:\$PATH\""
    ;;
esac

info "Done. Try:"
echo "  $INSTALL_DIR/$BIN_NAME --help"
echo "  $INSTALL_DIR/$BIN_NAME runtime detect"
