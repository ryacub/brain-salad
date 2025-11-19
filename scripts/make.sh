#!/bin/bash

# Quick build script with options

set -e

# Use the directory containing this script as the project directory
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY_PATH="$PROJECT_DIR/target/release/tm"

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
    echo "  -i, --incremental  Use incremental compilation (default)"
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
        -i|--incremental)
            # Already enabled by default, but we'll make sure
            echo "${BLUE}💡 Incremental compilation is enabled by default in .cargo/config.toml${NC}"
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

if [ "$VERBOSE" = true ]; then
    echo "${BLUE}🏗️  Building Telos Matrix CLI...${NC}"
fi

# Force clean if requested
if [ "$FORCE_CLEAN" = true ]; then
    if [ "$QUIET" != true ]; then
        echo "${YELLOW}🧹 Cleaning previous builds...${NC}"
    fi
    cargo clean
fi

# Check only
if [ "$CHECK_ONLY" = true ]; then
    if [ "$QUIET" != true ]; then
        echo "${BLUE}🔍 Checking compilation...${NC}"
    fi
    cargo check
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
    time cargo build
    BINARY_PATH="$PROJECT_DIR/target/debug/tm"
elif [ "$RELEASE_BUILD" = true ]; then
    if [ "$QUIET" != true ]; then
        echo "${BLUE}🚀 Building release version (optimized but slower)...${NC}"
    fi
    time cargo build --release
else
    # Default case - should be covered by defaults but just in case
    if [ "$QUIET" != true ]; then
        echo "${BLUE}🔧 Building development version (faster compilation)...${NC}"
    fi
    time cargo build
    BINARY_PATH="$PROJECT_DIR/target/debug/tm"
fi

# Check if build succeeded
if [ ! -f "$BINARY_PATH" ]; then
    echo "${RED}❌ Build failed! Binary not found at $BINARY_PATH${NC}"
    exit 1
fi

# Update symlink only for release builds
if [ "$DEV_BUILD" != true ]; then
    ln -sf "$PROJECT_DIR/target/release/tm" "$HOME/.local/bin/tm"
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
    cargo test -- --nocapture --show-output
    if [ "$QUIET" != true ]; then
        echo "${GREEN}✅ All tests passed!${NC}"
    fi
fi

# Quick verification
if [ "$QUIET" != true ] && [ "$DEV_BUILD" != true ]; then
    echo "${BLUE}🔍 Quick verification...${NC}"
    tm --version >/dev/null 2>&1
    if [ $? -eq 0 ]; then
        echo "${GREEN}✅ tm command is working!${NC}"
    else
        echo "${YELLOW}⚠️  tm command not available in PATH, but binary is built${NC}"
    fi
fi