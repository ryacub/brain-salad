# LLM Setup and Usage Guide

## Scripts Overview

The telos-idea-matrix includes several scripts to manage your LLM setup:

### setup-llm.sh
- **Purpose**: Complete setup including installation, starting the service, and pulling models
- **Usage**: `./setup-llm.sh`
- **What it does**:
  - Installs Ollama if not already installed
  - Starts the Ollama service
  - Pulls the mistral model
  - Sets up auto-start service (optional)
  - Verifies Claude CLI availability (optional)

### start-llm.sh
- **Purpose**: Starts the Ollama service if it's not running
- **Usage**: `./start-llm.sh`
- **When to use**: When you want to start the LLM service without doing a full setup

### stop-llm.sh
- **Purpose**: Stops the Ollama service
- **Usage**: `./stop-llm.sh`
- **When to use**: When you want to stop the LLM service

### update-llm.sh
- **Purpose**: Updates your models and Ollama installation
- **Usage**: `./update-llm.sh`
- **What it does**:
  - Updates Ollama binary (if installed via package manager)
  - Pulls latest versions of existing models
  - Tests model availability

## Usage with Telos Idea Matrix

Once your LLM is running, the telos-idea-matrix will automatically use AI analysis when:

1. You run `tm add --ai` for detailed analysis with AI
2. You run `tm add` without flags for basic rule-based scoring

## Troubleshooting

### If Ollama service won't start:
- Check if another instance is running: `ps aux | grep ollama`
- Look at logs: `cat /tmp/ollama.log`

### If models aren't being used:
- Ensure the service is running: `ps aux | grep ollama`
- Check available models: `ollama list`

### For Claude CLI integration:
- Ensure `claude` command is available in PATH
- Check configuration with: `claude --help`

## Model Recommendations

The system is configured to use `mistral` by default, but you can also use:
- `mistral-dolphin` for more conversational responses
- `llama2` for more general purpose analysis
- `codellama` for technical analysis

To use a different model, run: `ollama pull <model-name>`