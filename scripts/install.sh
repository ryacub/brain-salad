#!/usr/bin/env bash
set -e

# Brain-Salad Installation Script
# Usage: curl -sSL https://raw.githubusercontent.com/rayyacub/brain-salad/main/scripts/install.sh | bash

INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

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

# Clone first to read go.mod for version
echo "Cloning brain-salad..."
git clone --depth 1 https://github.com/rayyacub/brain-salad.git
cd brain-salad

# Get required Go version from go.mod (single source of truth)
REQUIRED_GO_VERSION=$(grep -E '^go [0-9]+\.[0-9]+' go.mod | awk '{print $2}')
echo "Required Go version: $REQUIRED_GO_VERSION"

# Install Go if not present or version too old
install_go() {
    local version="$1"
    echo "Downloading Go $version..."

    GO_TAR="go${version}.${OS}-${ARCH}.tar.gz"
    curl -sLO "https://go.dev/dl/${GO_TAR}"

    if [ ! -f "$GO_TAR" ]; then
        echo "Error: Failed to download Go $version"
        exit 1
    fi

    tar -xzf "$GO_TAR" -C "$TEMP_DIR"
    rm "$GO_TAR"

    export PATH="$TEMP_DIR/go/bin:$PATH"
    export GOROOT="$TEMP_DIR/go"
    export GOPATH="$TEMP_DIR/gopath"
    mkdir -p "$GOPATH"

    echo "Go $version ready."
}

if ! command -v go &> /dev/null; then
    install_go "$REQUIRED_GO_VERSION"
else
    CURRENT_GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    # Compare major.minor versions
    CURRENT_MAJOR=$(echo "$CURRENT_GO_VERSION" | cut -d. -f1)
    CURRENT_MINOR=$(echo "$CURRENT_GO_VERSION" | cut -d. -f2)
    REQUIRED_MAJOR=$(echo "$REQUIRED_GO_VERSION" | cut -d. -f1)
    REQUIRED_MINOR=$(echo "$REQUIRED_GO_VERSION" | cut -d. -f2)

    if [ "$CURRENT_MAJOR" -lt "$REQUIRED_MAJOR" ] || \
       ([ "$CURRENT_MAJOR" -eq "$REQUIRED_MAJOR" ] && [ "$CURRENT_MINOR" -lt "$REQUIRED_MINOR" ]); then
        echo "Go $CURRENT_GO_VERSION found, but $REQUIRED_GO_VERSION required."
        install_go "$REQUIRED_GO_VERSION"
    else
        echo "Go $CURRENT_GO_VERSION found (>= $REQUIRED_GO_VERSION required). OK."
    fi
fi

# Build
echo "Building brain-salad..."
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
