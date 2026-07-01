#!/usr/bin/env bash
set -euo pipefail

BIN_NAME="bqa"
INSTALL_DIR="${BQA_INSTALL_DIR:-$HOME/.local/bin}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHECK_GO_ONLY=0

info() { printf "\033[1;34m%s\033[0m\n" "$*"; }
warn() { printf "\033[1;33m%s\033[0m\n" "$*"; }
fail() { printf "\033[1;31m%s\033[0m\n" "$*" >&2; exit 1; }

usage() {
  cat <<'USAGE'
Usage: bash install.sh [--check-go]

Options:
  --check-go   Check the installed Go version and exit without building bqa.
  -h, --help   Show this help.
USAGE
}

minimum_go_version() {
  local version
  version="$(awk '$1 == "go" { print $2; exit }' "$SCRIPT_DIR/go.mod" 2>/dev/null || true)"
  if [[ -z "$version" ]]; then
    version="1.22"
  fi
  printf "%s\n" "$version"
}

normalize_go_version() {
  local value="$1"
  value="${value#go}"
  value="${value%%[!0-9.]*}"
  printf "%s\n" "$value"
}

installed_go_version() {
  local raw version
  raw="$(go version 2>/dev/null || true)"
  version="$(normalize_go_version "$(printf "%s\n" "$raw" | awk '{ print $3 }')")"
  if [[ -z "$version" ]]; then
    fail "Could not parse Go version from: $raw"
  fi
  printf "%s\n" "$version"
}

version_at_least() {
  local actual="$1"
  local minimum="$2"
  local a_major a_minor a_patch m_major m_minor m_patch

  IFS=. read -r a_major a_minor a_patch <<<"$actual"
  IFS=. read -r m_major m_minor m_patch <<<"$minimum"

  a_major="${a_major:-0}"
  a_minor="${a_minor:-0}"
  a_patch="${a_patch:-0}"
  m_major="${m_major:-0}"
  m_minor="${m_minor:-0}"
  m_patch="${m_patch:-0}"

  if (( 10#$a_major > 10#$m_major )); then return 0; fi
  if (( 10#$a_major < 10#$m_major )); then return 1; fi
  if (( 10#$a_minor > 10#$m_minor )); then return 0; fi
  if (( 10#$a_minor < 10#$m_minor )); then return 1; fi
  (( 10#$a_patch >= 10#$m_patch ))
}

check_go_version() {
  if ! command -v go >/dev/null 2>&1; then
    fail "Go is required for this installer. Install it first: brew install go"
  fi

  local minimum actual
  minimum="$(minimum_go_version)"
  actual="$(installed_go_version)"

  if ! version_at_least "$actual" "$minimum"; then
    fail "Go $minimum or newer is required. Found: $actual. Upgrade with: brew upgrade go"
  fi

  info "Go version OK: $actual (minimum $minimum)"
}

while (($#)); do
  case "$1" in
    --check-go)
      CHECK_GO_ONLY=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      fail "Unknown option: $1"
      ;;
  esac
done

check_go_version

if [[ "$CHECK_GO_ONLY" == "1" ]]; then
  info "Skipping install because --check-go was requested."
  exit 0
fi

info "Installing BQA-OS from local checkout..."

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
