#!/bin/bash

# Stop Script for Ollama LLM Service
# This script stops the Ollama service

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${YELLOW}⏹️  Stopping LLM service...${NC}"

# Function to check if ollama serve is running
is_ollama_running() {
    pgrep -f "ollama serve" >/dev/null
}

if is_ollama_running; then
    # Kill the ollama serve process
    pkill -f "ollama serve"
    
    # Wait a moment to ensure it stops
    sleep 2
    
    if ! is_ollama_running; then
        echo -e "${GREEN}✅ Ollama service stopped successfully${NC}"
    else
        echo -e "${YELLOW}⚠️  Ollama service might still be running${NC}"
        # Force kill if necessary
        pkill -9 -f "ollama serve" 2>/dev/null || true
        echo -e "${GREEN}✅ Ollama service force stopped${NC}"
    fi
else
    echo -e "${BLUE}ℹ️  Ollama service was not running${NC}"
fi

echo -e "${GREEN}✅ LLM service stopped${NC}"