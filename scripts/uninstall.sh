#!/usr/bin/env bash
set -e

INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY_NAME="tm"
LEGACY_DATA_DIR="$HOME/.telos"
MODERN_DATA_DIR="$HOME/.brain-salad"

echo "Uninstalling Brain-Salad..."

if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
    sudo rm "$INSTALL_DIR/$BINARY_NAME"
    echo "✓ Removed $INSTALL_DIR/$BINARY_NAME"
fi

# Ask about modern data directory
if [ -d "$MODERN_DATA_DIR" ]; then
    read -p "Remove data directory $MODERN_DATA_DIR? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf "$MODERN_DATA_DIR"
        echo "✓ Removed $MODERN_DATA_DIR"
    fi
fi

# Ask about legacy data directory
if [ -d "$LEGACY_DATA_DIR" ]; then
    read -p "Remove legacy data directory $LEGACY_DATA_DIR? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf "$LEGACY_DATA_DIR"
        echo "✓ Removed $LEGACY_DATA_DIR"
    fi
fi

echo "Uninstall complete"
