#!/bin/bash

# File watcher for automatic build on changes
# This script monitors Rust source files and automatically builds when changes are detected

set -e

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WATCH_DIR="$PROJECT_DIR/src"
BUILD_SCRIPT="$PROJECT_DIR/make.sh"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "${BLUE}🔍 Starting file watcher for automatic builds...${NC}"
echo "${BLUE}📁 Watching directory: $WATCH_DIR${NC}"

# Function to run build
run_build() {
    echo "${YELLOW}🔄 Change detected, building...${NC}"
    "$BUILD_SCRIPT" -d -q  # Build development version for faster builds, quiet output
    if [ $? -eq 0 ]; then
        echo "${GREEN}✅ Build completed successfully!${NC}"
    else
        echo "${RED}❌ Build failed!${NC}"
    fi
}

# Initial build
run_build

# Check if fswatch is available (macOS)
if command -v fswatch &> /dev/null; then
    echo "${BLUE}✅ Using fswatch for file monitoring${NC}"
    fswatch -o "$WATCH_DIR" --event-flags | while read -r events; do
        if [ "$events" != "0" ]; then
            run_build
        fi
    done
# Check if inotifywait is available (Linux)
elif command -v inotifywait &> /dev/null; then
    echo "${BLUE}✅ Using inotifywait for file monitoring${NC}"
    while inotifywait -r -e modify,create,delete "$WATCH_DIR" --include=".*\.rs$"; do
        run_build
    done
# Fallback to find with sleep (works on most systems)
else
    echo "${YELLOW}⚠️  Neither fswatch nor inotifywait found. Using fallback method.${NC}"
    echo "${YELLOW}⚠️  This method may be slower and less efficient.${NC}"
    
    # Store initial file modification times
    PREV_MOD_TIME=$(find "$WATCH_DIR" -name "*.rs" -exec stat -c "%Y %n" {} \; 2>/dev/null | sort | md5sum)
    
    while true; do
        sleep 2  # Check every 2 seconds
        CURRENT_MOD_TIME=$(find "$WATCH_DIR" -name "*.rs" -exec stat -c "%Y %n" {} \; 2>/dev/null | sort | md5sum)
        
        if [ "$PREV_MOD_TIME" != "$CURRENT_MOD_TIME" ]; then
            PREV_MOD_TIME="$CURRENT_MOD_TIME"
            run_build
        fi
    done
fi