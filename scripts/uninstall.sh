#!/usr/bin/env bash
set -e

INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY_NAME="tm"
DATA_DIR="$HOME/.telos"

echo "Uninstalling Telos Idea Matrix CLI..."

if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
    sudo rm "$INSTALL_DIR/$BINARY_NAME"
    echo "✓ Removed $INSTALL_DIR/$BINARY_NAME"
fi

# Ask about data directory
if [ -d "$DATA_DIR" ]; then
    read -p "Remove data directory $DATA_DIR? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf "$DATA_DIR"
        echo "✓ Removed $DATA_DIR"
    fi
fi

echo "Uninstall complete"
