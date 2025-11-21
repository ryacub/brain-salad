#!/usr/bin/env bash
set -e

# Telos Idea Matrix Installation Script
# Usage: curl -sSL https://raw.githubusercontent.com/ryacub/brain-salad/main/scripts/install.sh | bash

VERSION="${1:-latest}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY_NAME="tm"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Installing Telos Idea Matrix CLI...${NC}"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    armv7l) ARCH="arm" ;;
    *)
        echo -e "${RED}Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

# Build from source if Go is available
if command -v go &> /dev/null; then
    echo -e "${YELLOW}Go detected. Building from source...${NC}"

    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"

    git clone https://github.com/ryacub/brain-salad.git
    cd brain-salad

    if [ "$VERSION" != "latest" ]; then
        git checkout "$VERSION"
    fi

    echo "Building..."
    CGO_ENABLED=1 go build -o "$BINARY_NAME" ./cmd/cli

    echo "Installing to $INSTALL_DIR..."
    sudo mv "$BINARY_NAME" "$INSTALL_DIR/"
    sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"

    cd ..
    rm -rf "$TEMP_DIR"

    echo -e "${GREEN}✓ Installed successfully!${NC}"
    echo ""
    echo "Run 'tm --help' to get started"
    echo "Run 'tm completion bash' to enable shell completion"

else
    echo -e "${RED}Go not found. Please install Go 1.24+ or download pre-built binary.${NC}"
    echo "Visit: https://github.com/ryacub/brain-salad/releases"
    exit 1
fi

# Optional: Set up data directory
DATA_DIR="$HOME/.telos"
if [ ! -d "$DATA_DIR" ]; then
    echo ""
    read -p "Create data directory at $DATA_DIR? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        mkdir -p "$DATA_DIR"
        echo -e "${GREEN}✓ Created $DATA_DIR${NC}"
        echo "Run 'tm init' to set up your workspace"
    fi
fi
