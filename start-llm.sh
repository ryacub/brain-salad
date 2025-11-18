#!/bin/bash

# Quick Start Script for Ollama LLM Service
# This script starts the Ollama service if not running

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Starting LLM service...${NC}"

# Function to check if ollama serve is running
is_ollama_running() {
    pgrep -f "ollama serve" >/dev/null
}

if is_ollama_running; then
    echo -e "${GREEN}✅ Ollama service is already running${NC}"
    echo -e "${BLUE}💡 Telos Idea Matrix can use LLM analysis${NC}"
else
    echo -e "${YELLOW}🔌 Starting Ollama service...${NC}"
    
    # Kill any existing ollama serve processes to avoid conflicts
    pkill -f "ollama serve" 2>/dev/null || true
    
    # Start ollama in background
    nohup ollama serve >/tmp/ollama-startup.log 2>&1 &
    
    # Wait for the service to start
    sleep 5
    
    if is_ollama_running; then
        echo -e "${GREEN}✅ Ollama service started successfully${NC}"
        echo -e "${GREEN}💡 Telos Idea Matrix can now use LLM analysis${NC}"
        echo -e "${BLUE}📝 Log file: /tmp/ollama-startup.log${NC}"
    else
        echo -e "${RED}❌ Failed to start Ollama service${NC}"
        echo -e "${YELLOW}Check logs: cat /tmp/ollama-startup.log${NC}"
        echo -e "${YELLOW}You can manually start with: ollama serve${NC}"
        exit 1
    fi
fi

# Verify model availability
echo -e "${BLUE}🔍 Checking for available models...${NC}"
if ollama list | grep -q "mistral"; then
    echo -e "${GREEN}✅ Mistral model is available${NC}"
else
    echo -e "${YELLOW}⚠️  Mistral model not found. Run: ollama pull mistral${NC}"
    echo -e "${BLUE}💡 To install: ollama pull mistral${NC}"
fi

echo -e "${GREEN}🎉 Ready to use LLM with Telos Idea Matrix!${NC}"
echo -e "${GREEN}💡 Run: tm dump${NC}"