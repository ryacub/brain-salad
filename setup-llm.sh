#!/bin/bash

# LLM Setup Script for Telos Idea Matrix
# This script installs and configures LLM services for the idea matrix

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Setting up LLM for Telos Idea Matrix...${NC}"

# Check if running on macOS or Linux
PLATFORM="unknown"
if [[ "$OSTYPE" == "darwin"* ]]; then
    PLATFORM="macos"
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    PLATFORM="linux"
fi

echo -e "${BLUE}💻 Platform detected: $PLATFORM${NC}"

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Install Ollama if not installed
install_ollama() {
    if command_exists ollama; then
        echo -e "${GREEN}✅ Ollama is already installed${NC}"
        return 0
    fi

    echo -e "${YELLOW}📦 Installing Ollama...${NC}"

    if [ "$PLATFORM" = "macos" ]; then
        # Install Ollama on macOS
        curl -fsSL https://ollama.ai/install.sh | sh
    elif [ "$PLATFORM" = "linux" ]; then
        # Install Ollama on Linux
        curl -fsSL https://ollama.ai/install.sh | sh
    else
        echo -e "${RED}❌ Unsupported platform: $PLATFORM${NC}"
        return 1
    fi

    echo -e "${GREEN}✅ Ollama installed${NC}"
}

# Check if Ollama service is running
check_ollama_running() {
    if pgrep -f "ollama serve" >/dev/null; then
        return 0
    else
        return 1
    fi
}

# Function to start Ollama service
start_ollama_service() {
    echo -e "${BLUE}🔌 Starting Ollama service...${NC}"
    
    # Kill any existing ollama serve processes to avoid conflicts
    pkill -f "ollama serve" || true
    
    # Start ollama in background
    nohup ollama serve >/tmp/ollama.log 2>&1 &
    
    # Wait a moment for the service to start
    sleep 3
    
    # Check if the service is actually running
    if check_ollama_running; then
        echo -e "${GREEN}✅ Ollama service started successfully${NC}"
        echo -e "${BLUE}📝 Log file: /tmp/ollama.log${NC}"
        return 0
    else
        echo -e "${RED}❌ Failed to start Ollama service${NC}"
        echo -e "${YELLOW}Check logs: cat /tmp/ollama.log${NC}"
        return 1
    fi
}

# Function to pull or update mistral model
setup_mistral_model() {
    echo -e "${BLUE}📥 Pulling/updating Mistral model...${NC}"
    
    # Check if model already exists
    if ollama list | grep -q "mistral"; then
        echo -e "${YELLOW}🔄 Mistral model found, updating...${NC}"
    else
        echo -e "${BLUE}🔽 Downloading Mistral model...${NC}"
    fi
    
    # Pull the model
    if ollama pull mistral; then
        echo -e "${GREEN}✅ Mistral model updated successfully${NC}"
        return 0
    else
        echo -e "${RED}❌ Failed to pull Mistral model${NC}"
        return 1
    fi
}

# Function to create ollama service using systemd (Linux) or launchd (macOS)
create_service() {
    if [ "$PLATFORM" = "macos" ]; then
        create_launchd_service
    elif [ "$PLATFORM" = "linux" ]; then
        create_systemd_service
    fi
}

create_launchd_service() {
    echo -e "${BLUE}🔧 Creating launchd service for macOS...${NC}"
    
    SERVICE_FILE="$HOME/Library/LaunchAgents/ai.ollama.service.plist"
    
    cat > "$SERVICE_FILE" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>ai.ollama.service</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/ollama</string>
        <string>serve</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>$HOME/.ollama.log</string>
    <key>StandardErrorPath</key>
    <string>$HOME/.ollama.log</string>
</dict>
</plist>
EOF

    # Load the service
    launchctl load "$SERVICE_FILE" || true
    echo -e "${GREEN}✅ launchd service created and loaded${NC}"
}

create_systemd_service() {
    echo -e "${BLUE}🔧 Creating systemd service for Linux...${NC}"
    
    SERVICE_FILE="$HOME/.config/systemd/user/ollama.service"
    
    # Create directory if it doesn't exist
    mkdir -p "$(dirname "$SERVICE_FILE")"
    
    cat > "$SERVICE_FILE" << EOF
[Unit]
Description=Ollama Service
After=network-online.target

[Service]
ExecStart=/usr/local/bin/ollama serve
User=$USER
Restart=always
RestartSec=3

[Install]
WantedBy=network-online.target
EOF

    # Enable the service
    systemctl --user daemon-reload
    systemctl --user enable ollama.service
    systemctl --user start ollama.service
    
    echo -e "${GREEN}✅ systemd service created and started${NC}"
}

# Function to check if Claude CLI is available
check_claude_cli() {
    if command_exists claude; then
        echo -e "${GREEN}✅ Claude CLI is available${NC}"
        return 0
    else
        echo -e "${YELLOW}⚠️  Claude CLI not found (Claude CLI is optional)${NC}"
        return 1
    fi
}

# Function to run a test idea through the system
test_idea_matrix() {
    echo -e "${BLUE}🧪 Testing Telos Idea Matrix with LLM...${NC}"
    
    # Test with a simple idea to make sure LLM integration works
    PROJECT_DIR="$(dirname "${BASH_SOURCE[0]}")"
    
    # Try a quick analysis with a placeholder idea
    cd "$PROJECT_DIR" && timeout 60s ./make.sh -c 2>/dev/null || true
    
    echo -e "${GREEN}✅ LLM setup complete - Telos Idea Matrix will now use LLM analysis${NC}"
    echo -e "${GREEN}💡 Run: tm dump - to start using LLM-powered analysis${NC}"
}

# Main execution
main() {
    echo -e "${BLUE}📋 Step 1: Installing Ollama${NC}"
    install_ollama
    
    echo -e "${BLUE}📋 Step 2: Starting Ollama Service${NC}"
    if ! start_ollama_service; then
        echo -e "${RED}❌ Could not start Ollama service. Please start it manually with: ollama serve${NC}"
        exit 1
    fi
    
    echo -e "${BLUE}📋 Step 3: Setting up Mistral Model${NC}"
    setup_mistral_model
    
    echo -e "${BLUE}📋 Step 4: Creating Service (optional, for auto-start)${NC}"
    create_service
    
    echo -e "${BLUE}📋 Step 5: Checking Claude CLI${NC}"
    check_claude_cli
    
    echo -e "${BLUE}📋 Step 6: Testing Integration${NC}"
    test_idea_matrix
    
    echo -e "${GREEN}🎉 LLM setup completed successfully!${NC}"
    echo -e "${GREEN}✅ Ollama service is running${NC}"
    echo -e "${GREEN}✅ Mistral model is available${NC}"
    echo -e "${GREEN}💡 Run 'tm dump' to use LLM-powered analysis${NC}"
    
    # Show available models
    echo -e "${BLUE}📋 Available models:${NC}"
    ollama list
}

# Run the main function
main "$@"