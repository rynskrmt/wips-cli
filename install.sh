#!/bin/sh
# Based on common installer scripts

set -e

OWNER="rynskrmt"
REPO="wips-cli"
BINARY="wip"
FORMAT="tar.gz"
BINDIR="./bin"

usage() {
  echo "Usage: $0 [-b bindir] [-v version]"
  echo "  -b bindir  Directory to install the binary to (default: ./bin)"
  echo "  -v version Version to install (default: latest)"
  exit 1
}

while getopts "b:v:h" arg; do
  case "$arg" in
    b) BINDIR="$OPTARG" ;;
    v) VERSION="$OPTARG" ;;
    h) usage ;;
    *) usage ;;
  esac
done

# Detect OS and Arch
OS=$(uname -s)
ARCH=$(uname -m)

case "$OS" in
  Linux) OS="Linux" ;;
  Darwin) OS="Darwin" ;;
  *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

case "$ARCH" in
  x86_64) ARCH="x86_64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  i386) ARCH="i386" ;;
  *) echo "Unsupported Arch: $ARCH"; exit 1 ;;
esac

# Determine Version
if [ -z "$VERSION" ]; then
  # Fetch latest version from GitHub API
  LATEST_URL="https://api.github.com/repos/$OWNER/$REPO/releases/latest"
  VERSION=$(curl -s $LATEST_URL | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
  if [ -z "$VERSION" ]; then
    echo "Error: Could not determine latest version."
    exit 1
  fi
fi

# GoReleaser naming convention: wips-cli_Linux_x86_64.tar.gz
# Note: Ensure the naming matches .goreleaser.yaml
# name_template: {{ .ProjectName }}_{{ title .Os }}_{{ if eq .Arch "amd64" }}x86_64...
# result: wips-cli_Linux_x86_64.tar.gz

ASSET_NAME="${REPO}_${OS}_${ARCH}.${FORMAT}"
DOWNLOAD_URL="https://github.com/$OWNER/$REPO/releases/download/$VERSION/$ASSET_NAME"

echo "Installing $REPO $VERSION to $BINDIR..."

# Create bin directory
mkdir -p "$BINDIR"

# Download and extract
tmp=$(mktemp -d)
echo "Downloading $DOWNLOAD_URL..."
curl -sfL "$DOWNLOAD_URL" | tar xz -C "$tmp"

# Move binary
mv "$tmp/$BINARY" "$BINDIR/$BINARY"
chmod +x "$BINDIR/$BINARY"

# Verify
echo "Installed successfully to $BINDIR/$BINARY"
"$BINDIR/$BINARY" --version

rm -rf "$tmp"
