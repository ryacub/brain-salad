# Architecture Documentation

## Overview

Telos Idea Matrix is a Rust CLI application that captures ideas and evaluates them against personalized goals and strategies. The system is designed to be modular, extensible, and privacy-focused with local-first architecture.

## System Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         CLI Interface (Clap)                           │
│                         ┌──────────────────┐                           │
│                    ┌────┤  dump  │ review │                           │
│                    │    │ prune  │  link  │                           │
│                    │    └────────┬────────┘                           │
├────────────────────┼───────────┬┘──────────────────────────────────────┤
│  LAYER 1:          │           │                                       │
│  Request           │           ▼                                       │
│  Processing        │    ┌─────────────────┐                           │
│                    │    │ CommandHandler  │ (Process user input)      │
│                    │    │ ┌─────────────┐ │                           │
│                    │    │ │  Validation │ │ (Sanitize, check bounds) │
│                    │    │ └──────┬──────┘ │                           │
│                    │    └────────┼────────┘                           │
│                    │             │                                     │
├────────────────────┼─────────────┼──────────────────────────────────────┤
│  LAYER 2:          │             ▼                                     │
│  Business          │    ┌─────────────────────────────────────┐       │
│  Logic             │    │  Scoring Strategy (Pluggable)      │       │
│                    │    │ ┌─────────────────────────────────┐ │       │
│                    │    │ │ Trait: ScoringStrategy          │ │       │
│                    │    │ │ ┌──────────────────────────────┐ │ │       │
│                    │    │ │ │ TelosScoringStrategy::score  │ │ │       │
│                    │    │ │ │ - Goal alignment (40%)       │ │ │       │
│                    │    │ │ │ - Pattern detection (35%)    │ │ │       │
│                    │    │ │ │ - Strategic fit (25%)        │ │ │       │
│                    │    │ │ └──────────────────────────────┘ │ │       │
│                    │    │ └─────────────────────────────────┘ │       │
│                    │    └────────────────┬────────────────────┘       │
│                    │                     │                             │
├────────────────────┼──────────────┬──────┼──────────────────────────────┤
│  LAYER 3:          │              │      ▼                             │
│  Integration       │              │   ┌──────────────────┐            │
│                    │              │   │ AI Integration   │            │
│                    │              │   │ (Optional)       │            │
│                    │              │   │ ┌──────────────┐ │            │
│                    │              │   │ │ Ollama       │ │            │
│                    │              │   │ │ Circuit      │ │            │
│                    │              │   │ │ Breaker      │ │            │
│                    │              │   │ └──────┬───────┘ │            │
│                    │              │   │        │ (fail) ▼             │
│                    │              │   │    Rule-based│               │
│                    │              │   └──────────────┘            │
│                    │              │                             │
│                    │    ┌─────────┴────────────────────┐        │
│                    │    ▼                              ▼        │
│                    │   ┌───────────────────────┐ ┌──────────┐  │
│                    │   │ Configuration Module  │ │ Telos    │  │
│                    │   │ ┌─────────────────────┤ │ Parser   │  │
│                    │   │ │ ConfigPaths         │ │          │  │
│                    │   │ │ - env var           │ │ Extracts │  │
│                    │   │ │ - ~/.config/        │ │ - Goals  │  │
│                    │   │ │ - ./telos.md        │ │ - Strats │  │
│                    │   │ │ - Custom paths      │ │ - Patterns│ │
│                    │   │ └─────────────────────┤ │          │  │
│                    │   └──────┬────────────────┘ └──────────┘  │
│                    │          │                    │            │
├────────────────────┼──────────┼────────────────────┼──────────────┤
│  LAYER 4:          │          │                    │              │
│  Persistence       │          ▼                    │              │
│                    │   ┌──────────────────────┐   │              │
│                    │   │ Database Layer       │   │              │
│                    │   │ ┌──────────────────┐ │   │              │
│                    │   │ │ SQLx (async)     │ │   │              │
│                    │   │ │ ┌──────────────┐ │ │   │              │
│                    │   │ │ │ SQLite DB    │ │ │   │              │
│                    │   │ │ │ - Ideas      │ │ │   │              │
│                    │   │ │ │ - Links      │ │ │   │              │
│                    │   │ │ │ - Tags       │ │ │   │              │
│                    │   │ │ │ - Analysis   │ │ │   │              │
│                    │   │ │ └──────────────┘ │ │   │              │
│                    │   │ └──────────────────┘ │   │              │
│                    │   └──────────────────────┘   │              │
│                    │                               │              │
│                    │    (Reads telos.md from)     │              │
│                    │    ◄────────────────────────┘              │
│                    │                                            │
└────────────────────┴────────────────────────────────────────────┘
```

## Module Structure

### Core Modules

- **`main.rs`**: CLI entry point and application orchestration
- **`commands/`**: CLI command implementations (dump, review, prune, etc.)
- **`config.rs`**: Configuration loading and management
- **`scoring.rs`**: Scoring engine with Telos alignment logic
- **`telos.rs`**: Telos configuration parsing
- **`database_simple.rs`**: Database operations with connection pooling
- **`ai/`**: AI integration layer with Ollama support
- **`errors/`**: Comprehensive error handling system

### Error Handling Architecture

```rust
pub enum ApplicationError {
    Database(#[from] DatabaseError),
    Scoring(#[from] ScoringError),
    Validation(#[from] ValidationError),
    Security(#[from] SecurityError),
    AiService(#[from] CircuitBreakerError),
    // ... other error variants
}
```

## Key Design Patterns

### 1. Configuration Abstraction
The system uses a hierarchical configuration loading system that supports multiple sources with priority order:
1. Environment variable
2. Current directory file
3. User config file
4. Interactive wizard

### 2. Circuit Breaker Pattern
AI integration uses circuit breaker pattern to gracefully handle service outages:
- Opens when failure threshold is reached
- Closes after recovery timeout
- Falls back to rule-based analysis

### 3. Async Architecture
The entire system is built with async/await patterns for high performance:
- Non-blocking database operations
- Concurrent AI requests
- Proper cancellation support

### 4. Type Safety
Extensive use of Rust's type system with:
- Newtype patterns for validation
- Trait abstractions for flexibility
- Compile-time safety guarantees

## Data Flow

1. **User Input**: Command line arguments processed by Clap
2. **Configuration**: ConfigPaths loaded from environment/config files
3. **Telos Parsing**: Telos configuration parsed from user's file
4. **Scoring**: Idea evaluated against user's goals and patterns
5. **AI Enhancement**: Optional AI analysis with fallback
6. **Persistence**: Results stored in SQLite database
7. **Output**: Results formatted and displayed to user

## Security Considerations

- Input validation on all user inputs
- SQL injection prevention via SQLx query macros
- Path traversal prevention in file operations
- Structured logging with sensitive data sanitization
- Circuit breaker for external services

## Extension Points

The architecture supports several extension points:

1. **Scoring Strategies**: Implement custom scoring logic via traits
2. **AI Providers**: Add new LLM providers to the AI module
3. **Commands**: Add new CLI commands in the commands module
4. **Storage**: Swap SQLite for other databases