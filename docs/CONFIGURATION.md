# Configuration Guide

## Overview

Brain-Salad is personalized per user via a Telos configuration file. This file describes your:
- Goals and deadlines
- Strategies and focus areas
- Technology stack preferences
- Known failure patterns

## Setting Up Your Telos File

### Option 1: Environment Variable (Recommended for Docker/CI)

```bash
export TELOS_FILE=/path/to/your/telos.md
tm add "Your idea"
```

### Option 2: Current Directory

Place your `telos.md` in your working directory:

```bash
cd /my/project
cp /path/to/my/telos.md .
tm add "Your idea"
```

### Option 3: Configuration File

Create `~/.config/telos-matrix/config.toml`:

```toml
telos_file = "/path/to/your/telos.md"
data_dir = "~/.local/share/telos-matrix"
log_dir = "~/.cache/telos-matrix/logs"
```

## Telos File Format

Your `telos.md` should contain sections like:

```markdown
# My Telos

## Goals
- G1: [Goal 1] (Deadline: YYYY-MM-DD)
- G2: [Goal 2] (Deadline: YYYY-MM-DD)
- G3: [Goal 3] (Deadline: YYYY-MM-DD)
- G4: [Goal 4] (Deadline: YYYY-MM-DD)

## Strategies
- S1: [Strategy 1]
- S2: [Strategy 2]
- S3: [Strategy 3]
- S4: [Strategy 4]

## Stack
- Primary: [Your main tech stack]
- Secondary: [Secondary technologies]

## Failure Patterns
- Pattern 1: [Description]
- Pattern 2: [Description]
```

## Migration from Personal Setup

If you're starting fresh, create your own telos.md based on your goals:

```bash
cp docs/examples/telos-template.md ./my-telos.md
# Edit my-telos.md to match your goals and patterns
export TELOS_FILE=$(pwd)/my-telos.md
tm dump "My first idea"
```

## Configuration Priority Order

The system looks for configuration in this order:

1. `TELOS_FILE` environment variable
2. `./telos.md` in current directory
3. `~/.config/telos-matrix/config.toml`
4. Interactive setup wizard (if nothing else is found)

## Troubleshooting Configuration Issues

### Common Issues

1. **File not found errors**: Check that your TELOS_FILE path is correct and file exists
2. **Permission errors**: Ensure the telos.md file has read permissions
3. **Format errors**: Verify your telos.md follows the expected format

### Verifying Configuration

To check which configuration is being loaded:

```bash
RUST_LOG=info tm dump "test"
```

This will show which configuration source is being used.