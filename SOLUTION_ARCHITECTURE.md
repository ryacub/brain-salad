# Solution Architecture: GitHub-Ready Telos Idea Matrix

## Bird's Eye View

```
┌────────────────────────────────────────────────────────────────────┐
│                    USER DEPLOYMENT OPTIONS                        │
├────────────────────────────────────────────────────────────────────┤
│                                                                    │
│  LOCAL CLI                  DOCKER              CLOUD (Future)    │
│  ┌──────────┐               ┌──────────────┐    ┌──────────────┐  │
│  │ macOS    │               │ Docker Image │    │ Web UI /     │  │
│  │ Linux    │               │ compose up   │    │ REST API     │  │
│  │          │               │              │    │              │  │
│  │ Download │               │ Any OS       │    │ Sync devices │  │
│  │ binary   │               │ Works        │    │              │  │
│  │ or build │               │              │    │ (v0.2.0+)    │  │
│  └──────────┘               └──────────────┘    └──────────────┘  │
│                                                                    │
└────────────────────────────────────────────────────────────────────┘
                                │
                   All options use same codebase
                         (single version)
```

---

## Component Architecture

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
│  Persistence       │          │                    │              │
│                    │          ▼                    │              │
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

                           DATA FLOW

User Input
    │
    ▼
Command Handler (process)
    │
    ├─→ Validation (check input)
    │
    ├─→ Config Loader (where is telos.md?)
    │
    ├─→ Telos Parser (read goals/patterns)
    │
    ├─→ Scoring Strategy (evaluate idea)
    │    │
    │    ├─→ Try AI (Ollama)
    │    │   │ (fail) ▼
    │    │   Rule-based fallback
    │    │
    │    └─→ Pattern Detection
    │
    ├─→ Database (store/retrieve)
    │
    └─→ Display Results
        │
        ▼
    User sees scored idea
```

---

## Data Structures

### Core Idea Type
```rust
pub struct Idea {
    id: UUID,
    content: String,
    score: f32,
    patterns: Vec<Pattern>,
    tags: Vec<String>,
    relationships: Vec<IdeaLink>,
    created_at: DateTime,
    last_updated: DateTime,
}

pub struct Pattern {
    name: String,           // "context-switching", "perfectionism"
    severity: Severity,     // High, Medium, Low
    description: String,
}

pub enum IdeaLink {
    DependsOn(UUID),
    RelatedTo(UUID),
    Blocks(UUID),
}
```

### Configuration Types
```rust
pub struct ConfigPaths {
    telos_file: PathBuf,      // Where is user's goal document?
    data_dir: PathBuf,         // Where to store database?
    log_dir: PathBuf,          // Where to write logs?
}

pub struct TelosConfig {
    goals: Vec<Goal>,          // G1-G4: User's main objectives
    strategies: Vec<Strategy>, // S1-S4: How to achieve them
    stack: TechStack,          // Primary/secondary tech focus
    failure_patterns: Vec<PatternRule>, // Known traps
}

pub struct Goal {
    name: String,
    deadline: Date,
    priority: u8,
}
```

---

## Deployment Scenarios

### Scenario 1: Ray's Personal Use (Current)
```
Ray's MacBook
    │
    ├─ Rust installed
    ├─ Cargo build --release
    │
    ├─ TELOS_FILE=/Users/rayyacub/.../telos.md (env var)
    │
    └─ tm dump "new idea"
       │
       ├─ Load config → /Users/rayyacub/.../telos.md
       ├─ Load telos (Ray's goals)
       ├─ Score against Ray's metrics
       ├─ Store in ~/.local/share/telos-matrix/ideas.db
       │
       └─ Ray: "Great! This aligns with G1"
```

### Scenario 2: Friend's Laptop (After GitHub)
```
Friend's MacBook/Linux
    │
    ├─ Download binary from GitHub release
    │
    ├─ Copy friend's telos.md to current directory
    │
    ├─ TELOS_FILE=./telos.md (env var or default)
    │
    └─ tm dump "my startup idea"
       │
       ├─ Load config → ./telos.md
       ├─ Load telos (Friend's goals: ship startup, get funding, etc.)
       ├─ Score against Friend's metrics
       ├─ Store in ~/.local/share/telos-matrix/ideas.db
       │
       └─ Friend: "Perfect! This avoids my perfectionism trap"
```

### Scenario 3: Docker (Any OS)
```
Friend's Windows/Mac/Linux
    │
    ├─ docker-compose up
    │
    ├─ Mount telos.md:
    │   - ./telos.md → /config/telos.md (read-only)
    │
    ├─ Mount data volume:
    │   - telos-data → /data (persistent)
    │
    └─ docker-compose exec telos-matrix dump "idea"
       │
       ├─ Load config → /config/telos.md
       ├─ Load telos (Friend's goals)
       ├─ Score against Friend's metrics
       ├─ Store in /data/ideas.db (persists after container stops)
       │
       └─ Friend: "Works perfectly in Docker!"
```

### Scenario 4: Programmatic Use (Library)
```
Developer using TIM as library
    │
    ├─ cargo add telos-idea-matrix
    │
    ├─ use telos_idea_matrix::*
    │
    └─ fn my_app() {
         config = ConfigPaths::load()?
         telos = telos::load(&config.telos_file)?
         scorer = TelosScoringStrategy::new(telos)
         idea = Idea::new("...")
         score = scorer.score(&idea).await
       }
```

---

## Extension Points

### Adding Custom Scoring Strategy

For organizations with OKR framework instead of Telos:

```
1. User creates new strategy:
   src/my_scoring/mod.rs
   │
   └─ impl ScoringStrategy for OkrScoring {
       async fn score(&self, idea: &Idea) -> f32 { ... }
     }

2. Register in main.rs:
   let strategy: Box<dyn ScoringStrategy> =
       Box::new(OkrScoring::new(okr_config));

3. No changes to database, commands, or other modules
   → Custom strategy plugs in seamlessly
```

### Adding New Command

For "export to Notion" feature:

```
1. Create src/commands/export_notion.rs

2. Implement command logic:
   pub async fn export_notion(
       ideas: Vec<Idea>,
       notion_key: String,
   ) -> Result<()> { ... }

3. Register in main.rs CLI:
   #[derive(Subcommand)]
   enum Commands {
       Dump { ... },
       Review { ... },
       ExportNotion { notion_key: String }, ← NEW
   }

4. No changes to scoring, database, or config
```

### Adding Custom AI Provider

To use Claude instead of Ollama:

```
1. Create src/ai/claude_provider.rs
   impl AiProvider for ClaudeProvider { ... }

2. Update src/ai/mod.rs:
   pub enum AiProvider {
       Ollama(OllamaClient),
       Claude(ClaudeClient), ← NEW
   }

3. Rest of code calls AiProvider trait
   → Works with any provider
```

---

## Data Persistence Model

```
~/.local/share/telos-matrix/
├── ideas.db                    ← SQLite database
│   ├── ideas table
│   │   ├── id (UUID)
│   │   ├── content (text)
│   │   ├── score (float)
│   │   ├── patterns (JSON)
│   │   ├── tags (JSON)
│   │   ├── created_at
│   │   └── updated_at
│   │
│   ├── idea_links table
│   │   ├── source_id
│   │   ├── target_id
│   │   └── link_type
│   │
│   └── Indexes on: score, created_at, content

~/.cache/telos-matrix/
├── logs/
│   ├── app.log          ← Structured logs
│   └── errors.log       ← Error details

~/.config/telos-matrix/
└── config.toml          ← Optional custom config
```

---

## Build & Release Pipeline

```
Developer commits code
    │
    ├─ Trigger: git push origin main
    │
    ▼
GitHub Actions Workflow: test.yml
├─ Run: cargo test --all-features
├─ Run: cargo clippy
├─ Run: cargo fmt --check
│
└─ Result: ✅ Pass or ❌ Fail
   (Fail = PR blocked until fixed)

Developer tags release
    │
    ├─ git tag v0.2.0
    ├─ git push origin v0.2.0
    │
    ▼
GitHub Actions Workflow: release.yml
├─ Build for multiple targets:
│  ├─ Linux x86_64
│  ├─ Linux ARM64
│  ├─ macOS Intel
│  └─ macOS Apple Silicon
│
├─ Create release on GitHub
│
├─ Upload binaries
│
└─ Users download and use
   (or: docker pull ray/telos-matrix:v0.2.0)
```

---

## Security & Privacy Model

```
DATA LOCATION        SENSITIVITY     HANDLING
────────────────────────────────────────────
Telos file           High (personal) → Local only, user controls
Ideas database       High (personal) → Local SQLite, encrypted if user wants
Logs                 Medium          → Local, can delete anytime
AI requests          High (if remote) → Optional, Ollama is local-only
Configuration        Medium          → Plain text, no secrets stored
```

**Security Features:**
- ✅ All data stays local (default)
- ✅ No authentication (since local)
- ✅ No remote calls (unless user adds Ollama network)
- ✅ Input validation (XSS, SQL injection prevention)
- ✅ Error messages don't leak sensitive data
- ✅ Database supports encryption (SQLite native)

---

## Performance Characteristics

```
OPERATION              TIME        BOTTLENECK
─────────────────────────────────────────────
Load config            ~1ms        File I/O
Load telos file        ~5ms        YAML parsing
Score idea (no AI)     ~10ms       Pattern matching
Score idea (with AI)   ~2-5sec     Ollama inference
Review 1000 ideas      ~200ms      Database query + display
Prune ideas            ~500ms      Database transaction
```

**Optimizations applied:**
- Async I/O (Tokio) → No blocking
- Connection pooling → Reuse DB connections
- Lazy loading → Don't load all ideas into memory
- Streaming output → Large exports don't hit RAM limits
- Caching → Remember computed patterns

---

## Testing Architecture

```
TEST LEVELS           COVERAGE            TOOLS
────────────────────────────────────────────────
Unit Tests            Core functions      cargo test
├─ Config loading
├─ Scoring logic
├─ Pattern detection
└─ Error handling

Integration Tests     Full workflows      cargo test
├─ Config → Scoring → DB
├─ Database operations
└─ Command execution

System Tests          End-to-end          Manual + CI
├─ CLI commands
├─ Docker builds
└─ Release artifacts

Performance Tests     Speed/memory        cargo bench (future)
└─ Large datasets
```

---

## Version Compatibility

```
COMPONENT           STABLE?    SUPPORT
────────────────────────────────────────
Core scoring        ✅ Yes     Until v1.0
Config format       ✅ Yes     Migrations provided
Database schema     ✅ Yes     Migrations provided
CLI interface       ⚠️  Maybe  Might add commands
Scoring trait       ⚠️  Maybe  Can add methods (backwards compat)
Docker image        ✅ Yes     Each version tagged
```

**Compatibility guarantee:** Patch versions (0.1.0 → 0.1.1) never break config/database.

---

## Going Forward: Extensibility Roadmap

### v0.1.0 (Current Release)
- ✅ Basic idea capture/review
- ✅ Telos-based scoring
- ✅ Local SQLite storage
- ✅ Optional Ollama integration
- ✅ Docker support

### v0.2.0 (Future: Q1 2025)
- 🔲 Web UI for review
- 🔲 Multiple Telos frameworks (OKR, SMART, etc.)
- 🔲 Batch operations
- 🔲 Advanced analytics
- 🔲 Homebrew package

### v1.0.0 (Future: Stability)
- 🔲 Stable API
- 🔲 Zero breaking changes guarantee
- 🔲 Full test coverage
- 🔲 Performance benchmarks
- 🔲 Production hardening

### v2.0.0+ (Future: Advanced)
- 🔲 Device sync (with E2E encryption)
- 🔲 Team/collaborative ideas
- 🔲 Mobile companion app
- 🔲 Advanced AI features

---

**This architecture enables:**
- Personal use (Ray)
- Friend/family customization
- Enterprise extensions (custom scoring)
- Academic research (pluggable strategies)
- Commercial variants (proprietary scorer)

**All from one codebase.**
