# Telos Idea Matrix - Complete Documentation

## Project Overview

The Telos Idea Matrix is a sophisticated CLI tool that captures ideas and evaluates them against your personal Telos framework. It helps combat decision paralysis and context-switching by transforming scattered thoughts into mission-aligned action items with intelligent pattern detection and scoring.

## Current Status: Production-Ready System

The system has evolved far beyond the original MVP scope and is now a comprehensive, production-ready application with:

- ✅ Advanced CLI with 8+ commands
- ✅ Async Rust architecture with Tokio
- ✅ SQLite database with connection pooling
- ✅ AI integration with Ollama and circuit breakers
- ✅ Comprehensive error handling system
- ✅ Structured logging with tracing
- ✅ Health monitoring and metrics
- ✅ Pattern detection for behavioral traps
- ✅ Bulk operations for idea management
- ✅ Idea linking and relationship tracking
- ✅ Performance optimization and monitoring

## Architecture Overview

```
telos-idea-matrix/
├── src/
│   ├── main.rs                 # CLI entry point with clap
│   ├── commands/               # Command implementations
│   │   ├── dump.rs             # Idea capture and analysis
│   │   ├── analyze.rs          # Detailed analysis
│   │   ├── score.rs            # Quick scoring
│   │   ├── review.rs           # Browse ideas
│   │   ├── prune.rs            # Idea cleanup
│   │   └── bulk.rs             # Bulk operations
│   ├── database_simple.rs      # SQLite operations with pooling
│   ├── scoring.rs              # Scoring engine with Telos alignment
│   ├── patterns_simple.rs      # Pattern detection engine
│   ├── ai/                     # AI integration layer
│   ├── display.rs              # Terminal UI formatting
│   ├── background_tasks.rs     # Task supervision
│   ├── logging.rs              # Structured logging
│   ├── validation.rs           # Input validation
│   ├── types.rs                # Type definitions
│   ├── traits.rs               # Abstraction interfaces
│   └── errors/                 # Comprehensive error system
│       ├── mod.rs              # Main error types
│       ├── database.rs         # Database errors
│       ├── scoring.rs          # Scoring errors
│       ├── validation.rs       # Validation errors
│       ├── security.rs         # Security errors
│       └── circuit_breaker.rs  # Circuit breaker errors
├── migrations/
│   └── 001_initial.sql         # Database schema
└── Cargo.toml                  # Dependencies and configuration
```

## Commands

### `dump` - Capture Ideas
Quickly captures ideas with immediate analysis:

```bash
# Capture from command line
telos-matrix dump "Build AI tool to analyze hotel reviews"

# Interactive capture
telos-matrix dump --interactive

# Quick capture without analysis
telos-matrix dump --quick "Note for later"
```

### `analyze` - Detailed Analysis
Provides comprehensive analysis of captured or specified ideas:

```bash
# Analyze last captured idea
telos-matrix analyze --last

# Analyze specific idea
telos-matrix analyze "Improve hotel check-in process"
```

### `score` - Quick Scoring
Quickly evaluate an idea without saving:

```bash
telos-matrix score "Build mobile app for hotel staff"
```

### `review` - Browse Ideas
Browse and manage stored ideas:

```bash
# Review all ideas
telos-matrix review

# Review with filters
telos-matrix review --limit 20 --min-score 7.0

# Show ideas needing pruning
telos-matrix review --pruning
```

### `prune` - Manage Old Ideas
Clean up low-value ideas:

```bash
# Interactive pruning
telos-matrix prune

# Auto-prune with confirmation
telos-matrix prune --auto

# Dry run (preview removals)
telos-matrix prune --dry-run
```

### `bulk` - Mass Operations
Perform operations on multiple ideas at once:

```bash
# Tag multiple ideas
telos-matrix bulk tag "important" --limit 10 --min-score 8.0

# Archive low-value ideas
telos-matrix bulk archive --older-than 30 --max-score 4.0

# Export ideas to CSV
telos-matrix bulk export my-ideas.csv --limit 500

# Import from CSV
telos-matrix bulk import new-ideas.csv --default-score 5.0
```

### `link` - Idea Relationships
Manage relationships between ideas:

```bash
# Link two ideas
telos-matrix link add <idea1_id> <idea2_id> --type "depends-on"

# Show related ideas
telos-matrix link show <idea_id>

# List all relationships for an idea
telos-matrix link list <idea_id>

# Remove a relationship
telos-matrix link remove <relationship_id>
```

### `health` - System Status
Check system health and metrics:

```bash
# Get health status
telos-matrix health

# Get health in JSON format
telos-matrix health --format json
```

## Scoring Algorithm

The system uses a sophisticated 3-part scoring mechanism:

### 1. Mission Alignment (40% weight)
- AI Systems (1.5 points max)
- Shipping Focus (0.8 points max)
- Income Potential (0.5 points max)
- Domain Expertise (1.2 points max)

### 2. Anti-Challenge (35% weight) 
- Context-Switching (1.2 points max)
- Perfectionism (1.0 points max)
- Accountability (0.8 points max)
- Income Anxiety (0.5 points max)

### 3. Strategic Fit (25% weight)
- Stack Compliance (1.0 points max)
- Shipping Habit (0.8 points max)
- Public Accountability (0.4 points max)
- Revenue Testing (0.3 points max)

## Pattern Detection

The system detects these behavioral traps:
- Context-switching (new tech stacks, shiny objects)
- Perfectionism (scope creep, over-engineering)
- Procrastination (learning before building, consumption traps)
- Accountability avoidance (solo-only projects)
- Scope creep (comprehensive vs MVP thinking)

## AI Integration

The system features local-first AI integration with:
- Ollama support for local LLMs
- Circuit breaker pattern for resilience
- Fallback to rule-based analysis when AI unavailable
- Performance metrics for AI operations
- Configurable models (currently using Mistral Dolphin)

## Error Handling

Comprehensive error handling with:
- Structured error types using `thiserror`
- Detailed error context for debugging
- Graceful degradation when services fail
- Circuit breaker pattern for AI service protection
- Retries with exponential backoff

## Performance & Optimization

- Async architecture with Tokio runtime
- Database connection pooling with monitoring
- Circuit breaker for AI service resilience
- Performance monitoring with timing metrics
- Memory optimization with efficient data structures
- Graceful shutdown handling with signal support

## Personalization

The system adapts to your specific Telos configuration:
- Reads your personal `telos.md` file
- Applies your current goals and strategies
- Detects your documented failure patterns
- Adjusts scoring based on your domain expertise
- Learns from your previous decisions and outcomes

## Security & Validation

- Input validation with injection protection
- Path traversal detection and prevention
- SQL injection prevention
- Sanitized output rendering
- Secure configuration handling

## Future Iteration Strategy

Since this is a personal tool, future iterations will be guided by actual usage patterns:

### Immediate Actions
1. Monitor daily usage patterns
2. Identify commands used most/least frequently
3. Refine scoring algorithm based on actual results
4. Optimize frequently used workflows

### Medium-term Enhancements
1. Advanced relationship queries (dependency graphs)
2. Integration with calendar for deadline awareness
3. Performance trend analysis
4. Custom report generation

### Long-term Considerations
1. Machine learning for pattern detection improvement
2. Cross-device synchronization
3. Plugin system for extensions

## Installation & Setup

### Prerequisites
```bash
# Install Rust
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

# Install and start Ollama for AI features (optional)
curl -fsSL https://ollama.ai/install.sh | sh
ollama serve
```

### Building
```bash
# Build the project
cargo build --release

# Run the application
./target/release/telos-matrix --help
```

## Usage Recommendations

1. **Start with dumps**: Begin using `dump` command for immediate value
2. **Establish routine**: Use regularly to build habit of capturing ideas
3. **Monitor patterns**: Pay attention to pattern detection alerts
4. **Review regularly**: Use `review` command to revisit older ideas
5. **Clean periodically**: Use `prune` to remove low-value ideas

## Troubleshooting

### Common Issues
- **AI unavailable**: Install Ollama with a model (e.g., `ollama pull mistral`)
- **Database locked**: Check for other processes accessing the database
- **Performance slow**: The system performs heavy analysis by default, use `--quick` flag for faster operations

### Logs
- Detailed logs are available with `RUST_LOG=debug` environment variable
- Errors are logged with full context for debugging
- Performance metrics are available through health command

## Performance Metrics

The system tracks:
- Command execution times
- Database operation performance
- AI service response times
- Memory and CPU usage
- Error rates and recovery

## Conclusion

The Telos Idea Matrix is a comprehensive solution for transforming scattered ideas into actionable, mission-aligned tasks. It goes far beyond the original MVP with enterprise-grade features while maintaining simplicity for everyday use.

The system is production-ready and immediately valuable for anyone struggling with idea prioritization and context-switching behaviors.