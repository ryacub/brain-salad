#!/bin/bash

# Quick script to run cargo fix and cleanup common warnings
echo "Running cargo fix to address basic issues..."

cd /Users/rayyacub/Documents/CCResearch/telos-idea-matrix

# Run cargo fix to handle basic issues
cargo fix --allow-dirty --broken-code

# Run clippy to address warnings
cargo clippy --fix --allow-dirty --allow-staged

echo "Cleanup completed!"