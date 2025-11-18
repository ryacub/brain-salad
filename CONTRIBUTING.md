# Contributing to Telos Idea Matrix

Thank you for your interest in contributing to the Telos Idea Matrix! We welcome contributions from the community and are grateful for your time and effort. Whether you're fixing a bug, adding a feature, improving documentation, or sharing ideas, your contribution makes this tool better for everyone.

## Table of Contents

- [Ways to Contribute](#ways-to-contribute)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Code Standards](#code-standards)
- [Testing Requirements](#testing-requirements)
- [Pull Request Process](#pull-request-process)
- [Issue Reporting](#issue-reporting)
- [Commit Message Guidelines](#commit-message-guidelines)
- [Code of Conduct](#code-of-conduct)

## Ways to Contribute

### Code Contributions
- Fix bugs or implement new features
- Improve performance or optimize existing code
- Add new commands or extend CLI functionality
- Enhance AI integration and scoring algorithms

### Documentation
- Improve existing documentation
- Add code examples and tutorials
- Write guides for specific use cases
- Fix typos or clarify confusing sections

### Community Support
- Answer questions in discussions
- Share your Telos configuration examples
- Report bugs and suggest enhancements
- Review pull requests

### Testing
- Write unit tests or integration tests
- Test new features and report feedback
- Improve test coverage

## Getting Started

Before contributing, please:

1. **Read the documentation**: Familiarize yourself with the [README](./README.md) and project architecture
2. **Check existing issues**: Look for related issues or discussions before starting work
3. **Create an issue first**: For major changes, open an issue to discuss your approach before implementing
4. **Join the community**: Connect with other contributors on [Discord](https://discord.gg/example)

## Development Setup

### Prerequisites

- **Rust 1.75 or higher**: Install via [rustup](https://rustup.rs/)
- **SQLite**: Required for database functionality (typically pre-installed on most systems)
- **Git**: For version control

### Setup Instructions

1. **Fork and Clone the Repository**
   ```bash
   git clone https://github.com/YOUR_USERNAME/telos-idea-matrix.git
   cd telos-idea-matrix
   ```

2. **Install Dependencies**
   ```bash
   # This will download and compile all dependencies
   cargo build
   ```

3. **Set Up the Database**
   ```bash
   # Run migrations (creates necessary tables)
   cargo run -- --help  # First run creates the database
   ```

4. **Create a Test Telos Configuration**

   Create a `telos.md` file in the project root for testing:
   ```markdown
   # My Test Telos

   ## Goals
   - Build useful CLI tools
   - Learn Rust deeply

   ## Strategies
   - Start simple, iterate often
   - Write tests first

   ## Current Stack
   - Rust, SQLite, Tokio

   ## Failure Patterns
   - Context switching between too many tools
   - Perfectionism before shipping
   ```

5. **Run the Project Locally**
   ```bash
   # Run in development mode
   cargo run -- dump "Test idea"

   # Run with specific command
   cargo run -- review

   # Run with verbose logging
   RUST_LOG=debug cargo run -- dump "Test with logging"
   ```

6. **Run Tests**
   ```bash
   cargo test
   ```

### Development Dependencies

The project uses these key dependencies:

- **clap** (4.4): CLI argument parsing
- **sqlx** (0.7): Async database operations with SQLite
- **tokio** (1.35): Async runtime
- **serde** (1.0): Serialization/deserialization
- **anyhow** & **thiserror**: Error handling
- **reqwest**: HTTP client for AI integration
- **ollama-rs**: LLM integration (optional)

See [Cargo.toml](./Cargo.toml) for the complete dependency list.

## Code Standards

### Rust Style Guidelines

We follow the official Rust style guidelines. All code must pass formatting and linting checks before submission.

#### Formatting
```bash
# Format all code
cargo fmt

# Check formatting without modifying files
cargo fmt -- --check
```

#### Linting
```bash
# Run Clippy with default lints
cargo clippy

# Run Clippy with warnings as errors (CI requirement)
cargo clippy -- -D warnings
```

### Code Quality Guidelines

1. **Write Idiomatic Rust**
   - Use iterators instead of loops where appropriate
   - Leverage Rust's type system for safety
   - Prefer `Result` types over panics
   - Use `?` operator for error propagation

2. **Error Handling**
   - Use `anyhow::Result` for application errors
   - Use `thiserror` for custom error types
   - Provide meaningful error messages
   - Don't use `.unwrap()` or `.expect()` in production code (tests are OK)

3. **Documentation**
   - Document all public APIs with doc comments (`///`)
   - Include examples in doc comments where helpful
   - Use `//` for inline comments explaining complex logic
   - Run `cargo doc --open` to preview documentation

4. **Async Code**
   - Use `async/await` consistently
   - Don't block the async runtime with CPU-intensive work
   - Use `tokio::spawn` for concurrent operations
   - Properly handle cancellation and timeouts

5. **Database Code**
   - Use SQLx compile-time checked queries
   - Always use parameterized queries (never string interpolation)
   - Handle database errors gracefully
   - Write migrations for schema changes

### Naming Conventions

- **Modules**: snake_case (`telos_parser`, `scoring_engine`)
- **Structs**: PascalCase (`TelosConfig`, `ScoringEngine`)
- **Functions**: snake_case (`calculate_score`, `load_config`)
- **Constants**: SCREAMING_SNAKE_CASE (`DEFAULT_TIMEOUT_MS`)

### Import Organization

Organize imports in this order:
1. Standard library
2. External crates
3. Internal modules

```rust
use std::path::PathBuf;

use anyhow::Result;
use serde::{Deserialize, Serialize};

use crate::config::ConfigPaths;
use crate::scoring::ScoringEngine;
```

### Architecture Guidelines

- **Modular Design**: Keep modules focused on a single responsibility
- **Trait Usage**: Use traits for abstractions and testability
- **Dependency Injection**: Pass dependencies explicitly rather than using globals
- **Configuration**: Use the existing config system for new settings
- **Logging**: Use `tracing` crate for structured logging

## Testing Requirements

All contributions must include appropriate tests and pass the existing test suite.

### Running Tests

```bash
# Run all tests
cargo test

# Run tests with output
cargo test -- --nocapture

# Run specific test
cargo test test_name

# Run tests with logging
RUST_LOG=debug cargo test

# Run integration tests only
cargo test --test '*'
```

### Test Coverage

- **Unit Tests**: Test individual functions and modules in isolation
- **Integration Tests**: Test complete workflows in `tests/` directory
- **Property Tests**: Consider property-based testing for complex logic

### Writing Good Tests

1. **Test Structure**: Follow the Arrange-Act-Assert pattern
2. **Test Isolation**: Tests should not depend on each other
3. **Descriptive Names**: Use clear, descriptive test function names
4. **Test Data**: Use temporary directories for file-based tests (see `tempfile` crate)
5. **Mock External Services**: Don't rely on external APIs in tests

Example test:
```rust
#[tokio::test]
async fn test_idea_scoring_mission_alignment() {
    // Arrange
    let telos = create_test_telos();
    let idea = "Build a Rust CLI tool";

    // Act
    let score = score_idea(&telos, idea).await.unwrap();

    // Assert
    assert!(score >= 7.0);
    assert!(score <= 10.0);
}
```

## Pull Request Process

### Before Submitting

1. **Create a Feature Branch**
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/bug-description
   ```

2. **Make Your Changes**
   - Follow code standards
   - Write tests
   - Update documentation

3. **Verify Your Changes**
   ```bash
   # Format code
   cargo fmt

   # Run linter
   cargo clippy -- -D warnings

   # Run tests
   cargo test

   # Build the project
   cargo build --release
   ```

4. **Update Documentation**
   - Update README if adding new features
   - Add doc comments to new public APIs
   - Update CHANGELOG (if exists) with your changes

### Submitting the PR

1. **Push Your Branch**
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Create Pull Request**
   - Use a clear, descriptive title
   - Reference any related issues (e.g., "Fixes #123")
   - Provide a detailed description of changes
   - Include screenshots/examples if applicable

3. **PR Description Template**
   ```markdown
   ## Description
   Brief description of what this PR does

   ## Motivation
   Why is this change needed?

   ## Changes
   - List of changes made
   - Another change

   ## Testing
   How has this been tested?

   ## Related Issues
   Fixes #123
   ```

### Review Process

- Maintainers will review your PR within a few days
- Address feedback and requested changes
- Keep discussions respectful and constructive
- Once approved, a maintainer will merge your PR

### What Gets Reviewed

- Code quality and adherence to standards
- Test coverage and quality
- Documentation completeness
- Performance implications
- Security considerations
- Breaking changes (avoided when possible)

### CI/CD Pipeline Expectations

Your PR must pass:
- [ ] `cargo test` - All tests pass
- [ ] `cargo clippy --all-targets --all-features -- -D warnings` - No clippy warnings
- [ ] `cargo fmt --check` - Code is properly formatted
- [ ] `cargo build` - Builds successfully
- [ ] `docker build .` - Container builds successfully (if Dockerfile changed)

## Issue Reporting

### When to Create an Issue

- **Bug Reports**: Unexpected behavior or errors
- **Feature Requests**: New functionality or improvements
- **Questions**: Clarification on usage or behavior (consider Discussions first)
- **Documentation Issues**: Errors or gaps in documentation

### How to Write Good Issues

#### Bug Reports

Include the following information:
- **Description**: Clear description of the bug
- **Steps to Reproduce**: Exact steps to reproduce the issue
- **Expected Behavior**: What you expected to happen
- **Actual Behavior**: What actually happened
- **Environment**: OS, Rust version, project version
- **Logs**: Relevant error messages or logs (use `RUST_LOG=debug`)

Example:
```markdown
**Bug**: Scoring fails with empty telos.md

**Steps to Reproduce**:
1. Create empty telos.md file
2. Run `tm dump "test idea"`

**Expected**: Helpful error message
**Actual**: Panic with stack trace

**Environment**:
- OS: macOS 14.0
- Rust: 1.75.0
- Version: 0.1.0
```

#### Feature Requests

Include:
- **Problem Statement**: What problem does this solve?
- **Proposed Solution**: Your suggested approach
- **Alternatives**: Other solutions you've considered
- **Additional Context**: Examples, mockups, or use cases

### When to Create a PR Instead

Create a PR directly (without an issue) for:
- Typo fixes
- Documentation improvements
- Minor code cleanup
- Obvious bugs with simple fixes

For larger changes, always create an issue first to discuss the approach.

## Commit Message Guidelines

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification for clear and structured commit messages.

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- **feat**: New feature
- **fix**: Bug fix
- **docs**: Documentation changes
- **style**: Code style changes (formatting, no logic change)
- **refactor**: Code refactoring (no feature change or bug fix)
- **test**: Adding or updating tests
- **chore**: Maintenance tasks (dependencies, tooling)
- **perf**: Performance improvements

### Examples

```bash
feat(scoring): add weighted scoring for mission alignment

Implement a new scoring algorithm that weights mission alignment
at 40% of total score, with strategic fit at 25% and pattern
detection at 35%.

Closes #42
```

```bash
fix(database): handle connection pool exhaustion

Add retry logic with exponential backoff when connection pool
is exhausted. Improves reliability under high load.
```

```bash
docs(readme): add installation instructions for Windows

Add detailed Windows installation steps and troubleshooting
section for common issues.
```

### Best Practices

- Use imperative mood ("add feature" not "added feature")
- Keep subject line under 50 characters
- Capitalize subject line
- Don't end subject line with a period
- Separate subject from body with blank line
- Wrap body at 72 characters
- Explain *what* and *why*, not *how*

## Code of Conduct

### Our Standards

We are committed to providing a welcoming and inclusive environment:

- **Be Respectful**: Treat everyone with respect and kindness
- **Be Constructive**: Provide helpful feedback and suggestions
- **Be Patient**: Help newcomers learn and grow
- **Be Professional**: Keep discussions focused and professional
- **Be Open**: Welcome diverse perspectives and ideas

### Unacceptable Behavior

- Harassment, discrimination, or personal attacks
- Trolling, insulting, or derogatory comments
- Public or private harassment
- Publishing others' private information
- Other conduct inappropriate in a professional setting

### Reporting

If you experience or witness unacceptable behavior, please report it to the maintainers at ray@example.com. All reports will be reviewed and investigated promptly and fairly.

### Attribution

This Code of Conduct is adapted from the [Contributor Covenant](https://www.contributor-covenant.org/), version 2.1.

## Areas for Contribution

### Immediate Needs

- **Performance**: Improve scoring speed for large Telos files
- **AI Integration**: Add support for more LLM providers
- **UI**: Web interface for idea management and visualization
- **Documentation**: API reference for all commands
- **Testing**: More comprehensive integration tests

### Enhancement Ideas

- **Analytics Dashboard**: Visualize idea patterns over time
- **Multi-user Support**: Teams and organizations
- **API Server Mode**: REST API for web interfaces
- **Plugin System**: Extendable scoring algorithms
- **Import/Export**: JSON/YAML format support

### Bug Fixes

Check the [Issues](https://github.com/rayyacub/telos-idea-matrix/issues) page for reported bugs. Good candidates for first-time contributors include:
- Small documentation fixes
- Minor UI/UX improvements
- Error message clarifications
- Performance optimizations for edge cases

## Architecture Overview

### Core Components

```
main.rs          # CLI entry point & command routing
├── commands/    # CLI command implementations
│   ├── dump.rs  # Idea capture & analysis
│   ├── review.rs  # Idea browsing & management
│   └── ...      # Other commands
├── config/      # Configuration loading & management
├── scoring/     # Telos alignment scoring
├── telos/       # Telos parsing & processing
├── database/    # Data storage & retrieval
├── ai/          # AI integration layer
└── types/       # Shared data structures
```

### Key Patterns Used

- **Builder Pattern**: For complex object construction
- **Strategy Pattern**: For different scoring algorithms
- **Observer Pattern**: For event handling
- **Repository Pattern**: For database operations
- **Command Pattern**: For CLI command structure

### Error Handling Philosophy

- Use `anyhow::Result<T>` for application-level errors
- Use `thiserror` for domain-specific error types
- Provide context at the boundary of modules
- Don't panic in library code; return Result instead

## Getting Help

- **Questions**: Open a Discussion in the GitHub repo
- **Bugs**: File an issue with reproducible steps
- **Feature Requests**: Open an issue describing the use case
- **Direct Help**: Join our Discord server (link in README)

## Recognition

All contributors are recognized in the README and release notes. Major contributions may lead to maintainership opportunities.

Thank you for contributing to Telos Idea Matrix! Your efforts help make this tool better for everyone.
