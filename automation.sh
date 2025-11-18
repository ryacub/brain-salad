#!/bin/bash

# Comprehensive automation script for telos-idea-matrix
# This script handles automatic building on changes, development workflows, and releases

set -e

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD_SCRIPT="$PROJECT_DIR/make.sh"
BINARY_PATH="$PROJECT_DIR/target/release/tm"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

usage() {
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  dev-watcher     Start file watcher for development (fast builds)"
    echo "  git-auto-build  Setup Git hooks for automatic builds"
    echo "  release         Build and package for release"
    echo "  deploy          Deploy to production (if configured)"
    echo "  help            Show this help message"
    echo ""
}

# Function to run development build
dev_build() {
    echo "${YELLOW}🔄 Running development build...${NC}"
    "$BUILD_SCRIPT" -d -q  # Quick quiet dev build
    if [ $? -eq 0 ]; then
        echo "${GREEN}✅ Development build completed!${NC}"
    else
        echo "${RED}❌ Development build failed!${NC}"
    fi
}

# Function to run release build
release_build() {
    echo "${BLUE}🚀 Running release build...${NC}"
    "$BUILD_SCRIPT" -q  # Quiet release build
    if [ $? -eq 0 ]; then
        echo "${GREEN}✅ Release build completed!${NC}"
        echo "${CYAN}📦 Binary available at: $BINARY_PATH${NC}"
    else
        echo "${RED}❌ Release build failed!${NC}"
    fi
}

# Function to start file watcher
start_watcher() {
    echo "${BLUE}🔍 Starting file watcher for automatic builds...${NC}"
    echo "${BLUE}📁 Watching source directory: $PROJECT_DIR/src${NC}"
    echo "${YELLOW}💡 Press Ctrl+C to stop watching${NC}"

    # Store initial file modification times
    PREV_MOD_TIME=$(find "$PROJECT_DIR/src" -name "*.rs" -exec stat -c "%Y %n" {} \; 2>/dev/null | sort | md5sum)
    
    while true; do
        sleep 2  # Check every 2 seconds
        CURRENT_MOD_TIME=$(find "$PROJECT_DIR/src" -name "*.rs" -exec stat -c "%Y %n" {} \; 2>/dev/null | sort | md5sum)
        
        if [ "$PREV_MOD_TIME" != "$CURRENT_MOD_TIME" ]; then
            PREV_MOD_TIME="$CURRENT_MOD_TIME"
            dev_build
        fi
    done
}

# Function to setup Git hooks
setup_git_hooks() {
    HOOKS_DIR="$PROJECT_DIR/.git/hooks"
    
    if [ ! -d "$HOOKS_DIR" ]; then
        echo "${RED}❌ Git repository not found. Run 'git init' first.${NC}"
        exit 1
    fi
    
    # Create post-commit hook
    cat > "$HOOKS_DIR/post-commit" << 'EOF'
#!/bin/bash

# Git post-commit hook to automatically build the project
set -e

PROJECT_DIR="$(git rev-parse --show-toplevel)"
BUILD_SCRIPT="$PROJECT_DIR/automation.sh"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "${BLUE}🔄 Post-commit: Building project...${NC}"
"$BUILD_SCRIPT" dev-build

if [ $? -eq 0 ]; then
    echo "${GREEN}✅ Build successful after commit!${NC}"
else
    echo "${RED}❌ Build failed after commit!${NC}"
fi
EOF
    
    # Create post-merge hook
    cat > "$HOOKS_DIR/post-merge" << 'EOF'
#!/bin/bash

# Git post-merge hook to automatically build the project
set -e

PROJECT_DIR="$(git rev-parse --show-toplevel)"
BUILD_SCRIPT="$PROJECT_DIR/automation.sh"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "${BLUE}🔄 Post-merge: Building project after merge/pull...${NC}"
"$BUILD_SCRIPT" release

if [ $? -eq 0 ]; then
    echo "${GREEN}✅ Build successful after merge!${NC}"
else
    echo "${RED}❌ Build failed after merge!${NC}"
fi
EOF

    chmod +x "$HOOKS_DIR/post-commit" "$HOOKS_DIR/post-merge"
    echo "${GREEN}✅ Git hooks installed successfully!${NC}"
}

# Handle commands
case "${1:-help}" in
    "dev-watcher")
        start_watcher
        ;;
    "git-auto-build")
        setup_git_hooks
        ;;
    "release"|"-r")
        release_build
        ;;
    "dev-build"|"-d")
        dev_build
        ;;
    "deploy"|"-p")
        echo "${BLUE}🚀 Starting deployment process...${NC}"
        release_build
        if [ $? -eq 0 ]; then
            echo "${GREEN}✅ Deployment preparation complete!${NC}"
            # Add deployment-specific steps here if needed
        else
            echo "${RED}❌ Deployment preparation failed!${NC}"
            exit 1
        fi
        ;;
    "help"|"-h"|"--help")
        usage
        ;;
    *)
        echo "${RED}❌ Unknown command: $1${NC}"
        echo ""
        usage
        exit 1
        ;;
esac