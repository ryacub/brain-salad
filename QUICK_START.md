# Telos Idea Matrix - Quick Usage Guide

## Installation & Setup

### 1. First-time setup:
```bash
# Run the setup script to install everything
./setup-aliases.sh
```

### 2. Reload your shell configuration:
```bash
source ~/.zshrc  # or source ~/.bashrc
```

## Available Commands

### Core Commands
- `tm` - Main application
- `tmd "your idea"` - Quick idea dump
- `tma` - Analyze last idea
- `tmr` - Review all ideas
- `tm --help` - Show all available options

### Development Commands
- `tmb` - Build project (use `tmb -d` for development, `tmb -r` for release)
- `tmb -c` - Check compilation only (fastest)

### LLM Management Commands
- `tms` - Start LLM service
- `tmst` - Stop LLM service  
- `tmm` - Update LLM models

## Quick Start Examples

### Capture a new idea:
```bash
tmd "Build a tool that helps analyze hotel review sentiment"
```

### Review your ideas:
```bash
tmr
```

### Start LLM service (for AI analysis):
```bash
tms
tm dump  # Now will use AI analysis
```

## Pro Tips

1. **Use the interactive mode**: After dumping an idea, choose "Add another idea" to keep adding without restarting
2. **Check compilation quickly**: Use `tmb -c` for instant syntax checking
3. **Use development builds**: Use `tmb -d` for faster development iterations
4. **Watch for changes**: Run `./build-watcher.sh` to automatically build when files change

## Troubleshooting

- If `tm` command is not found: Make sure `~/.local/bin` is in your PATH
- If LLM isn't working: Run `tms` to start the service, then `tmm` to update models
- If aliases aren't working: Restart your terminal or run `source ~/.zshrc`