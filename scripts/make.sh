#!/bin/bash

# Quick build script with options for Go project

set -e

# Use the parent directory as the project root (scripts is a subdirectory)
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BINARY_PATH="$PROJECT_DIR/bin/tm"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help     Show this help message"
    echo "  -c, --check    Check compilation only (fastest option)"
    echo "  -t, --test     Run tests after build"
    echo "  -v, --verbose  Verbose output"
    echo "  -q, --quiet    Quiet output"
    echo "  -f, --force    Force clean build"
    echo "  -d, --dev      Build development binary (debug, faster)"
    echo "  -r, --release  Build release version (optimized)"
    echo ""
}

# Default options
CHECK_ONLY=false
RUN_TESTS=false
VERBOSE=false
QUIET=false
FORCE_CLEAN=false
DEV_BUILD=false
RELEASE_BUILD=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            exit 0
            ;;
        -c|--check)
            CHECK_ONLY=true
            shift
            ;;
        -t|--test)
            RUN_TESTS=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -q|--quiet)
            QUIET=true
            shift
            ;;
        -f|--force)
            FORCE_CLEAN=true
            shift
            ;;
        -d|--dev)
            DEV_BUILD=true
            RELEASE_BUILD=false  # Don't do both
            shift
            ;;
        -r|--release)
            RELEASE_BUILD=true
            DEV_BUILD=false  # Don't do both
            shift
            ;;
        *)
            echo "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Default to dev build if no specific build type is requested
if [ "$DEV_BUILD" = false ] && [ "$RELEASE_BUILD" = false ] && [ "$CHECK_ONLY" = false ]; then
    DEV_BUILD=true
    if [ "$QUIET" != true ]; then
        echo "${YELLOW}💡 No build type specified, defaulting to development build (faster)${NC}"
    fi
fi

# Change to project directory
cd "$PROJECT_DIR"

# Ensure bin directory exists
mkdir -p "$PROJECT_DIR/bin"

if [ "$VERBOSE" = true ]; then
    echo "${BLUE}🏗️  Building Telos Matrix CLI...${NC}"
fi

# Force clean if requested
if [ "$FORCE_CLEAN" = true ]; then
    if [ "$QUIET" != true ]; then
        echo "${YELLOW}🧹 Cleaning previous builds...${NC}"
    fi
    rm -rf "$PROJECT_DIR/bin"
    mkdir -p "$PROJECT_DIR/bin"
fi

# Check only
if [ "$CHECK_ONLY" = true ]; then
    if [ "$QUIET" != true ]; then
        echo "${BLUE}🔍 Checking compilation...${NC}"
    fi
    go build ./...
    if [ "$QUIET" != true ]; then
        echo "${GREEN}✅ Compilation check passed!${NC}"
    fi
    exit 0
fi

# Build
if [ "$DEV_BUILD" = true ]; then
    if [ "$QUIET" != true ]; then
        echo "${BLUE}🔧 Building development version (faster compilation)...${NC}"
    fi
    time go build -o "$BINARY_PATH" ./cmd/cli
elif [ "$RELEASE_BUILD" = true ]; then
    if [ "$QUIET" != true ]; then
        echo "${BLUE}🚀 Building release version (optimized but slower)...${NC}"
    fi
    time go build -ldflags="-s -w" -o "$BINARY_PATH" ./cmd/cli
else
    # Default case - should be covered by defaults but just in case
    if [ "$QUIET" != true ]; then
        echo "${BLUE}🔧 Building development version (faster compilation)...${NC}"
    fi
    time go build -o "$BINARY_PATH" ./cmd/cli
fi

# Check if build succeeded
if [ ! -f "$BINARY_PATH" ]; then
    echo "${RED}❌ Build failed! Binary not found at $BINARY_PATH${NC}"
    exit 1
fi

# Update symlink only for release builds
if [ "$DEV_BUILD" != true ]; then
    mkdir -p "$HOME/.local/bin"
    ln -sf "$BINARY_PATH" "$HOME/.local/bin/tm"
fi

if [ "$QUIET" != true ]; then
    echo "${GREEN}✅ Build successful!${NC}"
    if [ "$DEV_BUILD" = true ]; then
        echo "${BLUE}🔧 Development binary: $BINARY_PATH${NC}"
        echo "${YELLOW}💡 Use development build for iterative development${NC}"
    else
        echo "${BLUE}🚀 Release binary available as 'tm' command${NC}"
    fi
fi

# Run tests if requested
if [ "$RUN_TESTS" = true ]; then
    if [ "$QUIET" != true ]; then
        echo "${BLUE}🧪 Running tests...${NC}"
    fi
    go test -v ./...
    if [ "$QUIET" != true ]; then
        echo "${GREEN}✅ All tests passed!${NC}"
    fi
fi

# Quick verification
if [ "$QUIET" != true ] && [ "$DEV_BUILD" != true ]; then
    echo "${BLUE}🔍 Quick verification...${NC}"
    "$BINARY_PATH" --version >/dev/null 2>&1
    if [ $? -eq 0 ]; then
        echo "${GREEN}✅ tm command is working!${NC}"
    else
        echo "${YELLOW}⚠️  tm command not available in PATH, but binary is built${NC}"
    fi
fi