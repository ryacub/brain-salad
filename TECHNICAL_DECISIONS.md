# Technical Decisions & Trade-offs

This document explains key architectural decisions made for the GitHub-ready version of Telos Idea Matrix and the trade-offs involved.

---

## 1. Configuration Abstraction Pattern

### Decision: Hierarchical Configuration Loading

**Approach:**
1. Check environment variable (`TELOS_FILE`)
2. Check current directory (`./telos.md`)
3. Check config file (`~/.config/telos-matrix/config.toml`)
4. Error with helpful instructions

**Trade-offs:**

| Aspect | Trade-off |
|--------|-----------|
| **Flexibility** | 🟢 Users can deploy however they want |
| **Discoverability** | 🟡 Multiple options might confuse beginners |
| **Defaults** | 🟢 Works with sensible defaults |
| **Error messages** | 🟢 Clear instructions when missing |

**Why this approach:**
- **Docker**: Environment variable easy in containers
- **CLI**: Current directory works intuitively
- **Long-term**: Config file for persistent setup
- **Everyone wins**: Multiple use cases covered

**Alternative rejected:**
- Single hardcoded path → Too rigid, doesn't work for others
- Automatic discovery of ANY telos.md → Too fragile, confusing
- Required config wizard → Friction for new users

---

## 2. Pluggable Scoring Strategies

### Decision: Trait-based Strategy Pattern

**Approach:**
```rust
#[async_trait]
pub trait ScoringStrategy: Send + Sync {
    async fn score(&self, idea: &Idea) -> f32;
    async fn detect_patterns(&self, idea: &Idea) -> Vec<Pattern>;
}

// Implement once for Telos
pub struct TelosScoringStrategy { ... }

// Users can implement their own
pub struct CustomScoringStrategy { ... }
```

**Trade-offs:**

| Aspect | Trade-off |
|--------|-----------|
| **Extensibility** | 🟢🟢 Fully extensible, no limits |
| **Complexity** | 🟡 Slight abstraction overhead |
| **Performance** | 🟢 Dynamic dispatch minimal impact |
| **User learning** | 🟡 Rust users understand; others need doc |

**Why this approach:**
- **Future-proof**: Any scoring logic possible
- **Rust idiom**: Standard pattern developers expect
- **Minimal cost**: One trait, default implementation
- **Backwards compatible**: Ray's scoring stays unchanged

**Alternatives considered:**

| Alternative | Why rejected |
|-------------|-------------|
| **Hardcoded scoring** | Can't adapt to other frameworks (OKRs, SMART, etc.) |
| **Configuration-based DSL** | Overkill, harder to understand |
| **Simple enum dispatch** | Less flexible, harder to extend |
| **Plugin system with .so files** | Too complex for this scale |

---

## 3. Docker Multi-Stage Build

### Decision: Builder + Runtime Stages

**Approach:**
```dockerfile
FROM rust:1.75-slim as builder
  # Heavy: Full Rust toolchain, dependencies
  COPY Cargo.* .
  RUN cargo build --release

FROM debian:bookworm-slim
  # Light: Only binary + runtime dependencies
  COPY --from=builder /build/target/release/tm /usr/local/bin/tm
```

**Trade-offs:**

| Aspect | Trade-off |
|--------|-----------|
| **Image size** | 🟢 ~150MB (vs 2GB single-stage) |
| **Build time** | 🟠 ~3-5min first build (good cache) |
| **Complexity** | 🟢 Two lines explain it; still simple |
| **Maintenance** | 🟢 Single Dockerfile; minimal overhead |

**Why this approach:**
- **Size matters**: Users download image
- **Speed matters**: CI/CD builds frequently
- **Standard pattern**: Every serious project does this
- **Debuggability**: Can inspect builder stage if needed

**Alternatives rejected:**

| Alternative | Why rejected |
|-------------|-------------|
| **Single-stage** | 2GB image; 90% is unnecessary build tools |
| **Alpine Linux** | Smaller but musl issues with SQLite |
| **Distroless** | Harder to debug; less compatible |
| **Pre-built binaries only** | No flexibility; doesn't build on user systems |

---

## 4. Testing Strategy

### Decision: Integration Tests + CI Automation

**Approach:**
1. Integration tests for config loading (different scenarios)
2. Unit tests for scoring logic
3. GitHub Actions for CI
4. Docker build verification

**Trade-offs:**

| Aspect | Trade-off |
|--------|-----------|
| **Coverage** | 🟢 Core paths covered; edge cases tested |
| **Maintenance** | 🟡 Tests need updating with code |
| **Speed** | 🟢 Tests run in ~30sec in CI |
| **Effort** | 🟡 +30-40% development time |

**Why this approach:**
- **Configuration is critical**: One bad path breaks everything
- **Scoring is complex**: Pattern detection needs verification
- **Regression prevention**: Catch breakage early
- **User confidence**: CI badge shows project is maintained

**Test scope:**

| Category | Coverage | Why |
|----------|----------|-----|
| **Config loading** | 🟢 High | Affects all users |
| **Scoring logic** | 🟢 Medium-high | Core value; complex |
| **Database** | 🟡 Medium | Mostly SQLx responsibility |
| **AI integration** | 🟡 Low | Optional; has fallbacks |
| **CLI parsing** | 🟢 Low | Clap is well-tested |

**Test types:**

```
Unit Tests (20-30)
├── Configuration loading scenarios
├── Scoring algorithm correctness
├── Pattern detection accuracy
└── Error handling

Integration Tests (10-15)
├── Config file reading
├── Database operations
├── Full command execution
└── Docker image builds

CI Automation
├── Tests run on push
├── Clippy + fmt on PR
├── Docker builds on release
└── Binary distribution automated
```

---

## 5. Database Persistence & Local-First Design

### Decision: SQLite with No Server

**Approach:**
- Local SQLite database stored in `~/.local/share/telos-matrix/`
- No network calls, no authentication, no backend
- Async operations via SQLx + Tokio

**Trade-offs:**

| Aspect | Trade-off |
|--------|-----------|
| **Complexity** | 🟢 Simple; no server to manage |
| **Concurrency** | 🟡 One writer at a time (acceptable) |
| **Sync** | 🔴 Manual sync if multiple devices |
| **Backup** | 🟢 Just copy directory |
| **Scale** | 🟢 Handles 1000s of ideas easily |

**Why this approach:**
- **Fits the user**: Personal productivity tool
- **Privacy**: Data stays local, never sent
- **Reliability**: No network dependencies
- **Simplicity**: Users can inspect database directly

**Why NOT cloud/server-based:**
- Adds infrastructure complexity
- Introduces authentication/authorization
- Breaks offline-first philosophy
- Overkill for personal tool
- Data privacy concerns

**Data portability:**
- Export commands: CSV, JSON, Markdown
- Database is plain SQLite (open source tools)
- Configuration is TOML/YAML (human-readable)

---

## 6. AI Integration: Optional with Fallback

### Decision: Ollama for Local LLMs + Graceful Degradation

**Approach:**
```
User wants AI analysis
  ↓
Try to call Ollama
  ↓
  ├─ Success: Use AI analysis
  │
  └─ Timeout/Failed: Use rule-based scoring
```

**Trade-offs:**

| Aspect | Trade-off |
|--------|-----------|
| **User experience** | 🟢 Works with or without Ollama |
| **Complexity** | 🟡 Need fallback logic |
| **Performance** | 🟢 Doesn't block without AI |
| **Accuracy** | 🟡 Rules are okay, AI is better |

**Why this approach:**
- **Optional**: Not everyone wants to run Ollama
- **Works offline**: Rules work without network
- **Graceful**: System still useful without AI
- **Future-proof**: Can swap LLM providers

**Why Ollama specifically:**
- Local LLMs; private data
- Open source; can be self-hosted
- Easy setup (`ollama serve`)
- Supports multiple models (Mistral, Llama, etc.)

**Why NOT:**
- **OpenAI API**: Requires API key, costs money, phones home
- **Hardcoded LLM**: Forces dependency
- **No fallback**: System breaks if AI unavailable

---

## 7. Documentation Structure

### Decision: Multiple Docs at Different Depths

**Approach:**
```
README.md (Quick start + overview)
  ├─ CONFIGURATION.md (How to set up)
  ├─ ARCHITECTURE.md (Deep technical dive)
  ├─ API.md (Command reference)
  ├─ DOCKER_GUIDE.md (Container usage)
  └─ docs/plans/ (Implementation details)
```

**Trade-offs:**

| Aspect | Trade-off |
|--------|-----------|
| **Findability** | 🟢 Different docs for different questions |
| **Maintenance** | 🟡 More docs to keep in sync |
| **Completeness** | 🟢 Covers all use cases |
| **Discoverability** | 🟡 Users might not find right doc |

**Why layered approach:**
- **README**: Answer "What is this?" in 2 minutes
- **CONFIGURATION**: "How do I set this up?" detailed
- **ARCHITECTURE**: "How does it work?" for contributors
- **API**: Reference manual for commands
- **GUIDES**: Step-by-step for specific scenarios

**Why NOT single mega-document:**
- Too long; people won't read
- Different audiences want different info
- Easier to maintain separate docs
- Easier to search/reference

---

## 8. Version & Release Strategy

### Decision: Semantic Versioning + Automated Releases

**Approach:**
```
Tag format: v0.1.0 (semantic versioning)
  ↓
GitHub Actions builds binaries
  ↓
Attach to release page
  ↓
Users download pre-built or build from source
```

**Trade-offs:**

| Aspect | Trade-off |
|--------|-----------|
| **Automation** | 🟢 Releases fully automated |
| **Binaries** | 🟢 Don't need to build locally |
| **Versions** | 🟢 Clear semantic meaning |
| **Maintenance** | 🟡 Need CI/CD setup |

**Semantic versioning meaning:**

| Change | Version | Meaning |
|--------|---------|---------|
| New feature | v0.1.0 → v0.2.0 | Backwards compatible, can upgrade safely |
| Bug fix | v0.2.0 → v0.2.1 | Critical fix, upgrade strongly recommended |
| Breaking change | v0.x → v1.0.0 | API changed, might need config updates |

**Why this approach:**
- **Standard**: Every serious project does this
- **User safety**: Clear upgrade path
- **Predictability**: Users know what to expect

---

## 9. Error Handling Philosophy

### Decision: Structured Errors with Context

**Approach:**
```rust
pub enum ApplicationError {
    Config(#[from] ConfigError),
    Scoring(#[from] ScoringError),
    Database(#[from] DatabaseError),
    // ...
}

// Each error has helpful context and suggests fix
```

**Trade-offs:**

| Aspect | Trade-off |
|--------|-----------|
| **User experience** | 🟢 Errors explain what went wrong |
| **Debugging** | 🟢 Easy to trace problems |
| **Code** | 🟡 More error types to maintain |
| **Messages** | 🟢 Helpful, not cryptic |

**Why this approach:**
- **Context matters**: "Telos file not found at X" > "IO error"
- **User empowerment**: They can often fix themselves
- **Debugging**: Contributors can understand issues
- **Professional**: Shows polish and care

**Example error messages:**

❌ Bad:
```
Error: io error
```

✅ Good:
```
Error: Configuration Error
  Telos file not found at: /path/to/telos.md

  Please either:
  1. Set environment variable: export TELOS_FILE=/path/to/telos.md
  2. Place telos.md in current directory
  3. Create ~/.config/telos-matrix/config.toml with telos_file path

  See docs/CONFIGURATION.md for examples.
```

---

## 10. Deployment Targets: macOS/Linux + Docker

### Decision: Native Binaries + Container Image

**Approach:**
- **Primary**: Cargo build for macOS/Linux
- **Secondary**: Docker for consistency/other OS

**Trade-offs:**

| Platform | Approach | Trade-off |
|----------|----------|-----------|
| **macOS** | Native binary | Fastest; needs Rust installed OR pre-built binary |
| **Linux** | Native binary | Standard; works everywhere |
| **Windows** | Docker/WSL2 | Works but not primary; WSL2 adds friction |

**Why NOT Windows native:**
- Small user base for this tool
- WSL2 essentially Linux anyway
- Docker handles it well
- Can revisit later if demand

**Why NOT web-based:**
- Adds complexity (server, deployment)
- Local-first philosophy broken
- Frontend work required
- Can revisit as future enhancement

**Distribution method:**
1. **Developers**: `cargo build --release` + `cargo install`
2. **Non-Rust users**: Download pre-built from GitHub releases
3. **Docker users**: Build image locally or pull from registry
4. **Package managers**: Homebrew formula (future)

---

## 11. Organizational Philosophy: Convention Over Configuration

### Decision: Smart Defaults + Easy Overrides

**Examples:**

```bash
# Works with defaults (assumes telos.md in current dir)
tm dump "My idea"

# But can override everything
export TELOS_FILE=/custom/path/to/goals.md
tm dump "My idea"

# Docker: mounts data volume, persists across runs
docker-compose up
docker-compose exec telos-matrix dump "idea"
```

**Trade-offs:**

| Aspect | Trade-off |
|--------|-----------|
| **Ease of use** | 🟢 Works out of box |
| **Flexibility** | 🟢 Override everything if needed |
| **Discoverability** | 🟢 Defaults are sensible |
| **Learning curve** | 🟢 Minimal; understand as you go |

**Why this approach:**
- 80% of users: Default setup works
- 15% of users: One environment variable
- 5% of users: Full config file

---

## 12. Code Organization: Layered but Accessible

### Decision: Modular Structure without Over-engineering

**Structure:**
```
src/
├── main.rs          # Entry point; orchestration
├── config.rs        # Configuration loading
├── telos.rs         # Goal/Telos parsing
├── commands/        # CLI command implementations
├── scoring/         # Pluggable scoring logic
├── database.rs      # Data persistence
├── ai/              # Optional AI integration
├── errors.rs        # Error types
├── types.rs         # Core types (Idea, etc.)
└── ...
```

**Trade-offs:**

| Aspect | Trade-off |
|--------|-----------|
| **Clarity** | 🟢 Easy to understand |
| **Extensibility** | 🟢 Can extend without touching core |
| **Complexity** | 🟡 Not minimal, but reasonable |
| **Performance** | 🟢 No overhead from organization |

**Why this structure:**
- **Clear separation**: Each module has one job
- **Testability**: Easy to test in isolation
- **Contribution**: New contributors understand layout
- **Not over-engineered**: No unnecessary abstraction

---

## Summary: Balanced Trade-offs

| Principle | How we balance |
|-----------|--------------|
| **Simplicity ↔ Extensibility** | Simple CLI, pluggable internals |
| **Local ↔ Cloud** | Local-first, optional cloud later |
| **Automation ↔ Flexibility** | Smart defaults, easy overrides |
| **Features ↔ Scope** | Core features solid; extras planned |
| **Documentation ↔ Brevity** | Multiple docs; each focused |
| **Performance ↔ Maintainability** | Async where needed; not over-optimized |

---

## Decisions Still Open

### 1. Web UI (Future)
- Decision: **Postpone** until v0.2.0
- Reason: CLI is sufficient now; UI adds complexity
- Trigger: User demand or available bandwidth

### 2. Sync Across Devices
- Decision: **Not planned** for v0.1.0
- Reason: Adds server complexity; doesn't fit local-first
- Trigger: Community interest + clear use case

### 3. Team/Collaborative Telos
- Decision: **Possible future** via forks or git-based sharing
- Reason: Different from personal tool; different architecture
- Trigger: Demand from team use cases

### 4. Mobile App
- Decision: **Probably not** (out of scope)
- Reason: Rust + mobile is complex; web UI easier
- Trigger: Strong demand + dedicated resource

---

## How to Extend These Decisions

If you need to add features or change approach:

1. **Scoring logic**: Implement new `ScoringStrategy` trait
2. **Telos format**: Extend parsing in `src/telos.rs`
3. **Commands**: Add new command in `src/commands/`
4. **Storage**: Swap SQLite for PostgreSQL (update `src/database.rs`)
5. **AI**: Add new provider in `src/ai/`

All without breaking existing code.

---

**Last Updated**: November 17, 2025
