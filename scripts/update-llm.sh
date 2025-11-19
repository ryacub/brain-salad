#!/bin/bash

# Update Script for Ollama LLM Models
# This script updates the Ollama service and models

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔄 Updating LLM models and service...${NC}"

# Check if Ollama is installed
if ! command -v ollama &> /dev/null; then
    echo -e "${RED}❌ Ollama is not installed${NC}"
    echo -e "${YELLOW}Run: curl -fsSL https://ollama.ai/install.sh | sh${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Ollama is installed${NC}"

# Function to check if ollama serve is running
is_ollama_running() {
    pgrep -f "ollama serve" >/dev/null
}

# Check if service is running and warn if not
if ! is_ollama_running; then
    echo -e "${YELLOW}⚠️  Ollama service is not running${NC}"
    echo -e "${BLUE}💡 Start with: ./start-llm.sh${NC}"
    echo -e "${YELLOW}Some updates may require the service to be running${NC}"
else
    echo -e "${GREEN}✅ Ollama service is running${NC}"
fi

# Update Ollama (if installed via package manager on Linux)
echo -e "${BLUE}📦 Checking for Ollama service updates...${NC}"

# For macOS, check if installed via Homebrew
if [[ "$OSTYPE" == "darwin"* ]] && command -v brew &> /dev/null; then
    if brew list ollama &>/dev/null; then
        echo -e "${BLUE}Updating Ollama via Homebrew...${NC}"
        brew upgrade ollama || echo -e "${YELLOW}No updates available for Ollama${NC}"
    fi
fi

# Get current models and update them
echo -e "${BLUE}📋 Current models:${NC}"
ollama list

# Update the default model used by telos-idea-matrix
echo -e "${BLUE}🔄 Updating mistral model...${NC}"
if ollama pull mistral; then
    echo -e "${GREEN}✅ Mistral model updated successfully${NC}"
else
    echo -e "${RED}❌ Error updating mistral model${NC}"
fi

# Check for other commonly used models
COMMON_MODELS=("mistral" "llama2" "codellama" "phi" "gemma")

echo -e "${BLUE}🔄 Checking for other model updates...${NC}"
for model in "${COMMON_MODELS[@]}"; do
    if ollama list | grep -q "$model"; then
        echo -e "${BLUE}Updating $model...${NC}"
        if ollama pull "$model"; then
            echo -e "${GREEN}✅ $model updated${NC}"
        else
            echo -e "${YELLOW}⚠️  $model update failed${NC}"
        fi
    fi
done

# Show final model list
echo -e "${BLUE}📋 Updated models:${NC}"
ollama list

# Run a quick test to make sure service is working
echo -e "${BLUE}🧪 Testing model availability...${NC}"
if is_ollama_running; then
    echo -e "${BLUE}Sending test request...${NC}"
    if timeout 10s ollama run mistral "Hello, this is a test. Respond with 'Hello' only." >/tmp/ollama-test.txt 2>&1; then
        TEST_RESPONSE=$(head -c 20 /tmp/ollama-test.txt)
        echo -e "${GREEN}✅ Test successful: $TEST_RESPONSE...${NC}"
    else
        echo -e "${YELLOW}⚠️  Test timed out or failed (this is OK if service just started)${NC}"
    fi
else
    echo -e "${YELLOW}💡 Service not running, skipping test (run ./start-llm.sh first)${NC}"
fi

echo -e "${GREEN}🎉 LLM update completed!${NC}"
echo -e "${GREEN}💡 Models are updated and ready for use with Telos Idea Matrix${NC}"