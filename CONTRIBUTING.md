# Contributing to Brain-Salad

Thank you for your interest in contributing to Brain-Salad! We welcome contributions from the community and are grateful for your time and effort. Whether you're fixing a bug, adding a feature, improving documentation, or sharing ideas, your contribution makes this tool better for everyone.

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
- Answer questions in [GitHub Discussions](https://github.com/rayyacub/brain-salad/discussions)
- Share your configuration examples and scoring profiles
- Report bugs and suggest enhancements in [GitHub Issues](https://github.com/rayyacub/brain-salad/issues)
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
4. **Join the community**: Participate in [GitHub Discussions](https://github.com/rayyacub/brain-salad/discussions)

## Development Setup

### Prerequisites

- **Go 1.25.4 or higher**: Install via [official downloads](https://go.dev/dl/)
- **SQLite**: Required for database functionality (typically pre-installed on most systems)
- **Git**: For version control

### Setup Instructions

1. **Fork and Clone the Repository**
   ```bash
   git clone https://github.com/rayyacub/brain-salad.git
   cd brain-salad
   ```

2. **Install Dependencies**
   ```bash
   # This will download all dependencies
   go mod download
   ```

3. **Set Up the Database**
   ```bash
   # Run migrations (creates necessary tables)
   go run ./cmd/cli --help  # First run creates the database
   ```

4. **Create a Test Telos Configuration**

   Create a `telos.md` file in the project root for testing:
   ```markdown
   # My Test Telos

   ## Goals
   - Build useful CLI tools
   - Learn Go deeply

   ## Strategies
   - Start simple, iterate often
   - Write tests first

   ## Current Stack
   - Go, SQLite

   ## Failure Patterns
   - Context switching between too many tools
   - Perfectionism before shipping
   ```

5. **Run the Project Locally**
   ```bash
   # Run in development mode
   go run ./cmd/cli dump "Test idea"

   # Run with specific command
   go run ./cmd/cli review

   # Run with verbose logging
   LOG_LEVEL=debug go run ./cmd/cli dump "Test with logging"
   ```

6. **Run Tests**
   ```bash
   go test ./...
   ```

### Development Dependencies

The project uses these key Go modules:

- **github.com/spf13/cobra**: CLI command framework and argument parsing
- **modernc.org/sqlite**: Pure Go SQLite database driver
- **github.com/jmoiron/sqlx**: Database query extensions and utilities
- **github.com/olekukonko/tablewriter**: Terminal table formatting
- **github.com/fatih/color**: Terminal color output
- **github.com/sashabaranov/go-openai**: OpenAI API client
- **golang.org/x/sync/errgroup**: Goroutine error handling

See [go.mod](./go.mod) for the complete dependency list.

## Code Standards

### Go Style Guidelines

We follow the official Go style guidelines and best practices. All code must pass formatting and linting checks before submission.

#### Formatting
```bash
# Format all code
go fmt ./...

# Check formatting without modifying files (go fmt doesn't have --check, use diff)
gofmt -d .
```

#### Linting
```bash
# Run golangci-lint (requires installation)
golangci-lint run

# Run go vet for basic checks
go vet ./...

# Run golint for style checks (requires installation)
golint ./...
```

### Code Quality Guidelines

1. **Write Idiomatic Go**
   - Keep naming conventions (PascalCase for exported, camelCase for unexported)
   - Use interfaces for decoupling
   - Handle errors explicitly, don't panic
   - Prefer composition over inheritance

2. **Error Handling**
   - Use explicit error returns with context
   - Create custom error types with fmt.Errorf
   - Wrap errors to provide context
   - Provide meaningful error messages
   - Don't panic in production code (tests are OK)

3. **Documentation**
   - Document all public APIs with Go doc comments
   - Include examples in doc comments where helpful
   - Use `//` for inline comments explaining complex logic
   - Run `go doc .` to preview documentation

4. **Concurrency**
   - Use goroutines and channels appropriately
   - Don't block in critical paths
   - Use `context.Context` for cancellation and timeouts
   - Properly synchronize shared data with mutexes or channels

5. **Database Code**
   - Use prepared statements and parameterized queries
   - Always use parameterized queries (never string interpolation)
   - Handle database errors gracefully
   - Write migrations for schema changes

### Naming Conventions

- **Packages**: lowercase, single word preferred (`telos`, `scoring`, `analytics`)
- **Exported Types**: PascalCase (`TelosConfig`, `ScoringEngine`, `IdeaRepository`)
- **Unexported Types**: camelCase (`internalCache`, `scoreCalculator`)
- **Exported Functions**: PascalCase (`CalculateScore`, `LoadConfig`, `NewService`)
- **Unexported Functions**: camelCase (`calculateScore`, `loadConfig`, `parseInput`)
- **Exported Constants**: PascalCase (`DefaultTimeout`, `MaxRetries`)
- **Unexported Constants**: camelCase (`defaultBufferSize`, `maxConcurrency`)

### Import Organization

Organize imports in this order (goimports handles this automatically):
1. Standard library packages
2. External packages (third-party modules)
3. Internal packages (project modules)

```go
import (
    "context"
    "fmt"
    "path/filepath"

    "github.com/spf13/cobra"
    "github.com/jmoiron/sqlx"

    "github.com/rayyacub/telos-idea-matrix/internal/config"
    "github.com/rayyacub/telos-idea-matrix/internal/scoring"
)
```

### Architecture Guidelines

- **Modular Design**: Keep packages focused on a single responsibility
- **Interface Usage**: Use interfaces for abstractions and testability
- **Dependency Injection**: Pass dependencies explicitly rather than using globals
- **Configuration**: Use the existing config system for new settings
- **Logging**: Use the standard `log` package with configurable levels

## Testing Requirements

All contributions must include appropriate tests and pass the existing test suite.

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific test by name
go test -v ./... -run TestName

# Run tests with logging
LOG_LEVEL=debug go test -v ./...

# Run integration tests only
go test -v -tags=integration ./...

# Run tests with race detector
go test -race ./...

# Run tests with coverage
go test -cover ./...
```

### Test Coverage

- **Unit Tests**: Test individual functions and modules in isolation
- **Integration Tests**: Test complete workflows in `tests/` directory
- **Property Tests**: Consider property-based testing for complex logic

### Writing Good Tests

1. **Test Structure**: Follow the Arrange-Act-Assert pattern
2. **Test Isolation**: Tests should not depend on each other
3. **Descriptive Names**: Use clear, descriptive test function names
4. **Test Data**: Use temporary directories for file-based tests (see `t.TempDir()`)
5. **Mock External Services**: Don't rely on external APIs in tests

Example test:
```go
func TestIdeaScoringMissionAlignment(t *testing.T) {
    // Arrange
    telos := createTestTelos()
    idea := "Build a Go CLI tool"

    // Act
    score, err := scoreIdea(telos, idea)
    require.NoError(t, err)

    // Assert
    assert.GreaterOrEqual(t, score, 7.0)
    assert.LessOrEqual(t, score, 10.0)
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
   go fmt ./...

   # Run linter
   golangci-lint run

   # Run tests
   go test ./...

   # Build CLI binary
   go build ./cmd/cli

   # Build web binary
   go build ./cmd/web
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
- [ ] `go mod verify` - Dependency integrity check
- [ ] `go test -race ./...` - All tests pass with race detection
- [ ] `golangci-lint run` - No linting warnings
- [ ] `go test -tags=integration ./...` - Integration tests pass
- [ ] `go build ./cmd/cli` - CLI binary builds successfully
- [ ] `go build ./cmd/web` - Web binary builds successfully
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
- **Environment**: OS, Go version, project version
- **Logs**: Relevant error messages or logs (use `LOG_LEVEL=debug`)

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
- Go: 1.25.4
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

If you experience or witness unacceptable behavior, please report it via [GitHub Issues](https://github.com/rayyacub/brain-salad/issues). All reports will be reviewed and investigated promptly and fairly.

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
cmd/cli/main.go         # CLI entry point & command routing
├── cli/                # CLI command implementations
│   ├── dump.go         # Idea capture & analysis
│   ├── review.go       # Idea browsing & management
│   └── ...             # Other commands
├── config/             # Configuration loading & management
├── scoring/            # Telos alignment scoring
├── telos/              # Telos parsing & processing
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

- Return `error` as the last return value from functions
- Use `fmt.Errorf` with `%w` to wrap errors with context
- Create custom error types for domain-specific errors
- Provide context at the boundary of packages
- Don't panic in library code; return errors instead
- Check errors immediately after function calls

## Getting Help

- **Questions**: Open a Discussion in the GitHub repo
- **Bugs**: File an issue with reproducible steps
- **Feature Requests**: Open an issue describing the use case
- **Direct Help**: Join our Discord server (link in README)

## Recognition

All contributors are recognized in the README and release notes. Major contributions may lead to maintainership opportunities.

Thank you for contributing to Brain-Salad! Your efforts help make this tool better for everyone.
