#!/usr/bin/env bash
set -e

# Brain-Salad Installation Script
# Usage: curl -sSL https://raw.githubusercontent.com/rayyacub/brain-salad/main/scripts/install.sh | bash

INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
GO_VERSION="1.25.4"

echo "Installing Brain-Salad..."

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Create temp directory
TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"
trap "rm -rf $TEMP_DIR" EXIT

# Install Go if not present
if ! command -v go &> /dev/null; then
    echo "Go not found. Downloading Go $GO_VERSION..."

    GO_TAR="go${GO_VERSION}.${OS}-${ARCH}.tar.gz"
    curl -sLO "https://go.dev/dl/${GO_TAR}"
    tar -xzf "$GO_TAR"

    export PATH="$TEMP_DIR/go/bin:$PATH"
    export GOROOT="$TEMP_DIR/go"
    export GOPATH="$TEMP_DIR/gopath"
    mkdir -p "$GOPATH"

    echo "Go $GO_VERSION ready."
fi

# Clone and build
echo "Building brain-salad..."
git clone --depth 1 https://github.com/rayyacub/brain-salad.git
cd brain-salad
CGO_ENABLED=1 go build -o tm ./cmd/cli

# Install
echo "Installing to $INSTALL_DIR (may require sudo)..."
if [ -w "$INSTALL_DIR" ]; then
    mv tm "$INSTALL_DIR/"
else
    sudo mv tm "$INSTALL_DIR/"
fi

echo ""
echo "✓ Installed! Run 'tm init' to get started."
