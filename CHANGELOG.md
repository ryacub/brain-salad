# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Breaking Changes
- **LLM commands migrated to hierarchical structure:**
  - `tm llm-list` → `tm llm list`
  - `tm llm-config` → `tm llm config`
  - `tm llm-health` → `tm llm health`
  - The old flat commands have been removed

### Changed
- Install script now reads Go version from `go.mod` (single source of truth)
- Updated all documentation to reference `go.mod` for Go version requirements
- Added curl one-liner installation option to README
- Consolidated LLM analysis helpers into `internal/llm/analysis_helpers.go`
- Migrated internal logging from `fmt.Printf` to structured zerolog

### Added
- `.air-cli.toml` and `.air-api.toml` for hot reload development workflow
- `tm llm health` subcommand with `--watch` flag for continuous monitoring

### Fixed
- Fixed staticcheck SA5011 warnings in test files
- Fixed duplicate code between `cli/llm_helpers.go` and `cli/dump/llm.go`

### Removed
- Removed deprecated flat LLM commands (`llm-list`, `llm-config`, `llm-health`)
- Removed ~435 lines of duplicate code

## [2.0.1] - 2025-11-19

### Security
- Fixed SQL injection vulnerabilities in database queries
- Replaced insecure IP-based session tracking with SQLite-backed sessions
- Fixed rate limiter to properly handle proxy headers (X-Forwarded-For, X-Real-IP)
- Added proper session cookie security (Secure, HttpOnly, SameSite flags)

### Changed
- Reorganized project structure to follow standard Go conventions
  - Flattened structure by removing `/go` subdirectory
  - Moved `cmd/`, `internal/`, `pkg/`, `test/` to project root
  - Consolidated shell scripts into `/scripts` directory
  - Organized deployment files into `/deployments` structure
    - Docker files → `deployments/docker/`
    - Nginx configs → `deployments/nginx/`
    - Monitoring → `deployments/monitoring/`
  - Merged duplicate documentation and examples directories
  - Updated all import paths and file references
  - Updated GitHub workflows and Docker configurations
- Removed obsolete documentation (test reports, migration docs, Rust reference)
- Updated DEVELOPMENT.md to reflect new project structure
- Upgraded golangci-lint to v2 with updated configuration

### Fixed
- Fixed unchecked errors in error handling paths
- Reduced code duplication across the codebase
- Fixed critical documentation mismatch in ARCHITECTURE.md
- Fixed `.gitignore` patterns blocking `cmd/cli` directory
- Fixed missing `internal/cli` and `internal/telos` directories in git
- Fixed Docker workflow to use new Dockerfile location
- Fixed CI workflows for golangci-lint v2 compatibility

## [2.0.0] - 2024-11-19

### Changed
- **BREAKING:** Removed Rust implementation, transitioned to Go-only codebase
- Migrated all features to Go with 100% parity + enhancements
- Updated Docker configuration for Go builds
- Updated CI/CD to Go workflows only
- Restructured documentation (Rust docs → `docs/rust-reference/`)

### Added
- Interactive dump mode (`--interactive`) - Go exclusive feature
- Quick dump mode (`--quick`) - Go exclusive feature
- 5 LLM provider support (Ollama, OpenAI, Claude, Custom, Rule-based)
- Comprehensive test suite (18 packages, 11,000+ lines)
- Provider management with persistent configuration
- Semantic caching for LLM responses
- Quality metrics tracking

### Removed
- Rust source code (`src/`) - preserved in `archive/rust-implementation-2024` branch
- Rust build artifacts (`target/`)
- Rust configuration files (Cargo.toml, Cargo.lock, etc.)
- Rust CI workflows
- Rust benchmarks

### Migration
- Rust codebase archived in `archive/rust-implementation-2024` branch
- Rust documentation preserved in `docs/rust-reference/`
- See [MIGRATION_COMPLETE.md](docs/MIGRATION_COMPLETE.md) for full details

### Performance
- Build time improved: 5s (Go) vs 50s (Rust)
- Binary size: Comparable (~16MB)
- Runtime performance: Identical

### Breaking Changes for Developers
- Must use Go version specified in go.mod for development
- Build command changed: `go build` instead of `cargo build`
- Test command changed: `go test` instead of `cargo test`

### Breaking Changes for Users
- **None** - All CLI commands work identically

## [0.1.0] - 2025-11-18

### Added

#### Core Functionality
- **Idea Capture System**: Instant idea capture via `dump` command with automatic timestamping and organization
- **Telos-Aligned Scoring Engine**: Multi-dimensional scoring system evaluating ideas across three core dimensions:
  - Mission Alignment (40% weight) - Alignment with personal goals and mission
  - Anti-Pattern Detection (35% weight) - Detection of behavioral failure modes
  - Strategic Fit (25% weight) - Compatibility with current strategies and stack
- **Pattern Detection**: Automated detection of common failure patterns including:
  - Context-switching and shiny object syndrome
  - Perfectionism and over-engineering
  - Tutorial consumption vs. building (procrastination patterns)
  - Lack of accountability mechanisms
- **Intelligent Recommendations**: Context-aware recommendations system providing actionable guidance (Prioritize, Queue, Combine, Avoid, Break Down)

#### CLI Commands
- `dump` - Capture and analyze ideas with immediate scoring and storage
- `analyze` - Analyze ideas without storing them in the database
- `score` - Quick scoring without database persistence for rapid evaluation
- `review` - Browse and filter captured ideas with customizable filters
- `prune` - Automated and manual pruning of low-value ideas to prevent clutter
- `link` - Create and manage relationships between related ideas
- `bulk` - Perform batch operations on multiple ideas at once
- `analytics` - View statistics, trends, and insights about captured ideas
- `llm` - Manage local LLM integration and configuration

#### Configuration System
- **Multiple Configuration Methods**: Support for environment variables, config files, and interactive setup wizard
- **Flexible File Locations**: Automatic detection of `telos.md` in current directory, home directory, or custom paths
- **TOML Configuration**: User preferences stored in `~/.config/telos-matrix/config.toml`
- **Environment Variable Support**: `TELOS_FILE` for Docker and CI/CD environments
- **Interactive Setup Wizard**: Guided configuration for first-time users

#### Data Management
- **SQLite Database Backend**: Reliable, portable storage with no external database dependencies
- **Schema Migrations**: Automated database migrations using SQLx
- **CRUD Operations**: Complete create, read, update, delete operations for ideas
- **Data Export**: Export ideas to JSON, CSV, and Markdown formats
- **Relationship Tracking**: Store and query typed relationships between ideas (depends-on, related-to, blocks, etc.)

#### AI Integration
- **Ollama Integration**: Optional local LLM integration for enhanced analysis
- **Hybrid Analysis**: Combines rule-based scoring with AI-powered insights
- **Circuit Breaker Pattern**: Resilient AI integration with automatic fallback to rule-based analysis
- **Model Flexibility**: Support for multiple Ollama models (llama2, llama3, mistral, etc.)
- **Graceful Degradation**: Full functionality maintained when AI is unavailable

#### Docker Support
- **Pre-built Docker Images**: Multi-architecture images available on GitHub Container Registry
- **Docker Compose Configuration**: Ready-to-use compose setup for easy deployment
- **Volume Mounting**: Persistent data storage with volume mounts
- **Cross-platform Compatibility**: Images for linux/amd64 and linux/arm64

#### CI/CD Pipeline
- **GitHub Actions Workflows**: Automated testing, building, and deployment
- **Automated Testing**: Continuous integration with test suite execution on every push
- **Multi-platform Builds**: Automated builds for Linux, macOS, and Windows
- **Container Publishing**: Automatic Docker image builds and publishing to GHCR
- **Release Automation**: Automated release creation with cross-platform binaries

#### Development Infrastructure
- **Comprehensive Test Suite**: Unit tests, integration tests, and end-to-end tests
- **Benchmarking Suite**: Performance benchmarks for scoring and pattern detection
- **Logging System**: Structured logging with configurable verbosity using tracing/tracing-subscriber
- **Error Handling**: Robust error handling with detailed error messages using anyhow and thiserror
- **Type Safety**: Compile-time guarantees with SQLx's type-checked queries

#### User Experience
- **Rich Terminal Output**: Colored output with formatted tables and clear visual hierarchy
- **Interactive Prompts**: User-friendly prompts using dialoguer for multi-step operations
- **Progress Indicators**: Real-time feedback for long-running operations
- **Helpful Error Messages**: Clear, actionable error messages with suggested fixes
- **Comprehensive Help Text**: Detailed help documentation for all commands with examples

#### Documentation
- **Complete README**: Comprehensive project overview with quick start guide
- **Installation Guide**: Multiple installation methods (Docker, Cargo, pre-built binaries, source)
- **Configuration Documentation**: Detailed guide for configuration options and file formats
- **Docker Guide**: Advanced Docker usage patterns and best practices
- **Contributing Guide**: Guidelines for contributors including code style and testing requirements
- **Product Requirements Document**: Complete PRD with technical architecture and roadmap
- **Solution Architecture**: Detailed system design documentation
- **Technical Decisions**: ADR-style documentation of key architectural decisions

#### Project Management Features
- **Tag System**: Organize ideas with custom tags and filter by tags
- **Status Tracking**: Track idea lifecycle (pending, in-progress, completed, archived)
- **Linking System**: Connect related ideas with typed relationships
- **Bulk Operations**: Process multiple ideas simultaneously with batch commands
- **Trend Analysis**: Visualize patterns in ideation over time
- **Time-based Filtering**: Review ideas by date ranges (last 7 days, last month, etc.)
- **Score-based Filtering**: Filter ideas by minimum/maximum score thresholds

### Changed
- N/A (initial release)

### Deprecated
- N/A (initial release)

### Removed
- N/A (initial release)

### Fixed
- N/A (initial release)

### Security
- N/A (initial release)

---

## Project Links

- **Repository**: [https://github.com/ryacub/telos-idea-matrix](https://github.com/ryacub/telos-idea-matrix)
- **Docker Images**: [ghcr.io/rayyacub/telos-idea-matrix](https://github.com/ryacub/telos-idea-matrix/pkgs/container/telos-idea-matrix)
- **Issue Tracker**: [GitHub Issues](https://github.com/ryacub/telos-idea-matrix/issues)
- **Documentation**: [docs/](./docs/)

---

## Version History

[2.0.1]: https://github.com/ryacub/brain-salad/releases/tag/v2.0.1
[2.0.0]: https://github.com/ryacub/brain-salad/releases/tag/v2.0.0
[0.1.0]: https://github.com/ryacub/telos-idea-matrix/releases/tag/v0.1.0
