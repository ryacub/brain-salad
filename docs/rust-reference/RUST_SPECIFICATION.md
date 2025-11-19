# Rust Implementation Reference

This document describes the behavior of the Rust implementation to ensure the Go migration maintains feature parity.

**Last Updated:** 2025-11-19

---

## Table of Contents

1. [Data Models](#data-models)
2. [Scoring Algorithm](#scoring-algorithm)
3. [Telos Parser](#telos-parser)
4. [Pattern Detection](#pattern-detection)
5. [Database Schema](#database-schema)
6. [CLI Commands](#cli-commands)
7. [Test Cases](#test-cases)

---

## Data Models

### Core Type Wrappers

The Rust implementation uses newtype wrappers for type safety. These should be replicated in Go:

#### IdeaId

```rust
pub struct IdeaId(String)
```

- Wraps a string UUID or any valid ID (minimum 8 characters)
- Generated using `uuid::Uuid::new_v4().to_string()`
- Validation: Must be valid UUID or >= 8 characters
- Methods: `new()`, `as_str()`, `into_inner()`, `generate()`, `is_valid()`

#### Score

```rust
pub struct Score(f64)
```

- Range: 0.0 to 10.0
- Validation: Enforced at creation with `new()`
- Methods:
  - `is_priority()`: score >= 8.0
  - `is_good()`: score >= 6.0
  - `is_avoid()`: score < 4.0
  - `recommendation()`: Returns Recommendation enum based on score
- Display format: `{:.1}` (one decimal place)

#### Recommendation Enum

```rust
pub enum Recommendation {
    Priority,   // >= 8.5
    Good,       // >= 7.0
    Consider,   // >= 5.0
    Avoid,      // < 5.0
}
```

**Display Format:**
- Priority: "🔥 PRIORITIZE NOW"
- Good: "✅ GOOD ALIGNMENT"
- Consider: "⚠️ CONSIDER LATER"
- Avoid: "🚫 AVOID FOR NOW"

### Analysis Structures

#### Score Structure (from scoring.rs)

```rust
pub struct Score {
    pub mission: MissionScores,
    pub anti_challenge: AntiChallengeScores,
    pub strategic: StrategicScores,
    pub raw_score: f64,
    pub final_score: f64,
    pub recommendation: Recommendation,
    pub scoring_details: Vec<String>,
    pub explanations: HashMap<String, String>,
}
```

#### MissionScores

```rust
pub struct MissionScores {
    pub domain_expertise: f64,  // 0-1.2 points max
    pub ai_alignment: f64,      // 0-1.5 points max
    pub execution_support: f64, // 0-0.8 points max
    pub revenue_potential: f64, // 0-0.5 points max
    pub total: f64,             // max 4.0 points
}
```

#### AntiChallengeScores

```rust
pub struct AntiChallengeScores {
    pub context_switching: f64, // 0-1.2 points max
    pub rapid_prototyping: f64, // 0-1.0 points max
    pub accountability: f64,    // 0-0.8 points max
    pub income_anxiety: f64,    // 0-0.5 points max
    pub total: f64,             // max 3.5 points
}
```

#### StrategicScores

```rust
pub struct StrategicScores {
    pub stack_compatibility: f64,   // 0-1.0 points max
    pub shipping_habit: f64,        // 0-0.8 points max
    pub public_accountability: f64, // 0-0.4 points max
    pub revenue_testing: f64,       // 0-0.3 points max
    pub total: f64,                 // max 2.5 points
}
```

### Database Models

#### StoredIdea

```rust
pub struct StoredIdea {
    pub id: String,
    pub content: String,
    pub raw_score: Option<f64>,
    pub final_score: Option<f64>,
    pub patterns: Option<Vec<String>>,
    pub recommendation: Option<String>,
    pub analysis_details: Option<String>,
    pub created_at: DateTime<Utc>,
    pub reviewed_at: Option<DateTime<Utc>>,
    pub status: IdeaStatus,
}
```

#### IdeaStatus Enum

```rust
pub enum IdeaStatus {
    Active,
    Archived,
    Deleted,
}
```

String values: "active", "archived", "deleted"

#### IdeaRelationship

```rust
pub struct IdeaRelationship {
    pub id: String,
    pub source_idea_id: String,
    pub target_idea_id: String,
    pub relationship_type: RelationshipType,
    pub created_at: DateTime<Utc>,
}
```

#### RelationshipType Enum

```rust
pub enum RelationshipType {
    DependsOn,
    RelatedTo,
    PartOf,
    Parent,
    Child,
    Duplicate,
    Blocks,
    BlockedBy,
    SimilarTo,
}
```

String values: "depends_on", "related_to", "part_of", "parent", "child", "duplicate", "blocks", "blocked_by", "similar_to"

---

## Scoring Algorithm

### Formula

```
raw_score = mission.total + anti_challenge.total + strategic.total
final_score = (raw_score / 10.0) * 10.0  // Scale to 0-10
```

**Weight Distribution:**
- Mission Alignment: 4.0 points (40%)
- Anti-Challenge: 3.5 points (35%)
- Strategic Fit: 2.5 points (25%)
- **Total:** 10.0 points

### Mission Alignment Scoring (4.0 points max)

#### 1. Domain Expertise (1.2 points max)

Measures how well the idea leverages existing skills.

**Scoring Ranges:**
- **0.90-1.20**: Uses 80%+ existing skills
- **0.60-0.89**: Uses 50-79% existing skills
- **0.30-0.59**: Uses 30-49% existing skills
- **0.00-0.29**: Requires mostly new skills

**Algorithm:**
```rust
let match_ratio = matching_skills / total_skills

if match_ratio >= 0.8 {
    0.9 + ((match_ratio - 0.8) * 0.75)  // 0.9-1.2 range
} else if match_ratio >= 0.5 {
    0.6 + ((match_ratio - 0.5) * 0.967) // 0.6-0.89 range
} else if match_ratio >= 0.3 {
    0.3 + ((match_ratio - 0.3) * 0.967) // 0.3-0.59 range
} else {
    match_ratio * 1.033                  // 0.0-0.29 range
}
```

#### 2. AI Alignment (1.5 points max)

Measures how central AI is to the idea.

**Scoring Ranges:**
- **1.20-1.50**: Core product IS AI automation/systems
- **0.80-1.19**: AI is a significant component
- **0.40-0.79**: AI is auxiliary or optional
- **0.00-0.39**: Minimal or no AI component

**Keywords:**
- Core AI (1.2-1.5): "AI agent", "AI system", "automation pipeline", "build AI"
- Significant (0.8-1.19): "integrate AI", "using GPT", "powered by AI"
- Auxiliary (0.4-0.79): Generic "AI" mentions, "smart feature"

#### 3. Execution Support (0.8 points max)

Measures shipping timeline and MVP clarity.

**Scoring Ranges:**
- **0.65-0.80**: Clear deliverable within 30 days
- **0.45-0.64**: Deliverable within 60 days
- **0.25-0.44**: Longer timeline (90+ days)
- **0.00-0.24**: Learning-focused, no concrete deliverable

**Keywords:**
- Fast (0.65-0.8): "MVP", "30 days", "1 month", "prototype"
- Moderate (0.45-0.64): "60 days", "2 months", "basic version"
- Slow (0.25-0.44): "90 days", "comprehensive"
- Learning (0.0-0.24): "learn before", "study", "tutorial"

#### 4. Revenue Potential (0.5 points max)

Measures monetization clarity.

**Scoring Ranges:**
- **0.40-0.50**: Clear monetization model ($1K-$2.5K/month target)
- **0.25-0.39**: Plausible monetization ($500-$1K/month)
- **0.10-0.24**: Speculative monetization
- **0.00-0.09**: No clear revenue path

**Keywords:**
- High (0.4-0.5): "subscription", "SaaS", "$1000+", "recurring revenue"
- Medium (0.25-0.39): "freelance", "$500-$1K"
- Low (0.0-0.09): "ads", "free", "donation"

### Anti-Challenge Scoring (3.5 points max)

#### 1. Context Switching (1.2 points max)

Measures tech stack continuity.

**Scoring Ranges:**
- **0.95-1.20**: Uses 90%+ current stack
- **0.65-0.94**: Uses 70-89% current stack
- **0.30-0.64**: Requires 50%+ new stack elements
- **0.00-0.29**: Complete stack switch

**Penalty Keywords (0.0-0.29):**
- "Rust", "JavaScript", "TypeScript", "React", "Flutter", "mobile app", "game development"

#### 2. Rapid Prototyping (1.0 points max)

Measures MVP timeline.

**Scoring Ranges:**
- **0.80-1.00**: MVP in 1-2 weeks; inherently iterative (SaaS, automation)
- **0.55-0.79**: MVP in 3-4 weeks
- **0.25-0.54**: Requires 6+ weeks
- **0.00-0.24**: Perfection-dependent (content creation, courses)

#### 3. Accountability (0.8 points max)

Measures external accountability.

**Scoring Ranges:**
- **0.65-0.80**: Paying customers or public commitments
- **0.45-0.64**: Strong accountability structure
- **0.20-0.44**: Weak accountability
- **0.00-0.19**: No external accountability

**Keywords:**
- Strong (0.65-0.8): "customer", "client", "pre-order", "cohort"
- Moderate (0.45-0.64): "accountability partner", "public building"
- Weak (0.2-0.44): "social media", "personal goal"

#### 4. Income Anxiety (0.5 points max)

Measures time to first revenue.

**Scoring Ranges:**
- **0.40-0.50**: First revenue within 30 days
- **0.25-0.39**: First revenue within 60 days
- **0.10-0.24**: First revenue 90+ days
- **0.00-0.09**: Revenue 6+ months away

### Strategic Fit Scoring (2.5 points max)

#### 1. Stack Compatibility (1.0 points max)

Measures flow state potential.

**Scoring Ranges:**
- **0.80-1.00**: Enables 4+ hour flow sessions
- **0.55-0.79**: Allows 2-3 hour focus blocks
- **0.25-0.54**: Requires frequent context switching
- **0.00-0.24**: Inherently fragmented work

#### 2. Shipping Habit (0.8 points max)

Measures reusability for future projects.

**Scoring Ranges:**
- **0.65-0.80**: Creates reusable systems/code
- **0.45-0.64**: Some reusable components
- **0.20-0.44**: Minimal reusability
- **0.00-0.19**: Purely one-off effort

**Keywords:**
- High (0.65-0.8): "reusable", "library", "module", "framework", "system"
- Partial (0.45-0.64): "pattern", "transferable knowledge"
- One-off (0.0-0.19): "one-off", "unique", "custom", "bespoke"

#### 3. Public Accountability (0.4 points max)

Measures validation speed.

**Scoring Ranges:**
- **0.32-0.40**: Validate in 1-2 weeks (landing page, calls)
- **0.22-0.31**: Validation in 3-4 weeks
- **0.10-0.21**: Requires 6-8 weeks
- **0.00-0.09**: Requires 2+ months or full product

#### 4. Revenue Testing (0.3 points max)

Measures scalability potential.

**Scoring Ranges:**
- **0.24-0.30**: SaaS/product model; serves multiple customers
- **0.16-0.23**: Hybrid model; some leverage
- **0.08-0.15**: Service-based; limited leverage
- **0.00-0.07**: Pure time-for-money consulting

### Recommendation Thresholds

```rust
match final_score {
    s if s >= 8.5 => Recommendation::Priority,   // 🔥 PRIORITIZE NOW
    s if s >= 7.0 => Recommendation::Good,       // ✅ GOOD ALIGNMENT
    s if s >= 5.0 => Recommendation::Consider,   // ⚠️ CONSIDER LATER
    _ => Recommendation::Avoid,                  // 🚫 AVOID FOR NOW
}
```

---

## Telos Parser

### Expected Format

The Telos parser expects a Markdown file with the following structure:

```markdown
# Telos

**Last Updated:** YYYY-MM-DD

## PROBLEMS

### P1: Problem Title
Description of the problem.

**Why this matters:** Explanation

### P2: Next Problem
...

## MISSIONS

### M1: Mission Title
Description of the mission.

**Specific actions:**
- Action 1
- Action 2

## GOALS

### G1: Goal Title
- **What:** Description
- **Metric:** Measurable metric
- **Deadline:** YYYY-MM-DD
- **Why:** Reason

## CHALLENGES

### C1: Challenge Title
Description of the challenge.

**Evidence:** Supporting evidence

## STRATEGIES

### S1: The "One Stack, One Month" Rule
**The Rule:** Description

**Implementation:**
- November 2025: Python + LangChain + OpenAI
- December 2025: Next stack

**Why this works:** Explanation

## CURRENT STACK

(Extracted from S1 Implementation)
```

### Parsing Behavior

#### extract_current_stack()

Looks for Strategy S1 ("One Stack, One Month") and extracts the current month's stack from the Implementation section.

**Fallback:**
```rust
vec![
    "python", "langchain", "openai", "gpt", "api",
    "streamlit", "web app"
]
```

#### extract_domain_keywords()

Searches content for domain expertise keywords:
- Hotel/hospitality: "hotel", "hospitality", "hilton"
- Mobile: "mobile", "android", "app", "application"
- Software: "software", "development", "programming", "tech"

Returns: `Vec<String>` of matched keywords

#### parse_for_scoring()

Returns `TelosConfig`:

```rust
pub struct TelosConfig {
    pub current_stack: Vec<String>,
    pub domain_keywords: Vec<String>,
    pub income_deadline: String,        // From G1 deadline
    pub active_goals: Vec<String>,       // Goal titles
    pub active_strategies: Vec<String>,  // Strategy IDs (S1, S2, etc.)
    pub challenges: Vec<String>,         // Challenge titles
}
```

### Section Extraction

Uses regex patterns to find subsections:
- Problems: `^### P\d+:`
- Missions: `^### M\d+:`
- Goals: `^### G\d+:`
- Challenges: `^### C\d+:`
- Strategies: `^### S\d+:`

---

## Pattern Detection

### PatternType Enum

```rust
pub enum PatternType {
    ContextSwitching,        // 🔄
    Perfectionism,           // ⚡
    Procrastination,         // 🕰️
    AccountabilityAvoidance, // 👤
    ScopeCreep,             // 📏
}
```

### PatternMatch Structure

```rust
pub struct PatternMatch {
    pub pattern_type: PatternType,
    pub severity: Severity,
    pub matches: Vec<String>,
    pub message: String,
    pub suggestion: Option<String>,
}
```

### Severity Enum

```rust
pub enum Severity {
    Critical,  // 🔴
    High,      // 🟠
    Medium,    // 🟡
    Low,       // 🟢
    Positive,  // ✅
}
```

### Detection Rules

#### Context-Switching Detection

**Negative (High):**
- Keywords: "rust", "javascript", "react", "flutter", "swift"
- Message: "Context-switching risk detected"
- Suggestion: "Focus on current stack (Python + LangChain + OpenAI)"

**Positive:**
- Matches current stack keywords
- Message: "Staying focused on current tech stack"

#### Perfectionism Detection

**Negative (High):**
- Keywords: "comprehensive", "complete"
- Message: "Scope creep risk - over-engineering detected"
- Suggestion: "Define v1 scope and postpone advanced features"

#### Procrastination Detection

**Negative (Critical):**
- Pattern: "learn" + ("before" OR "then")
- Message: "Consumption trap - learning before building"
- Suggestion: "Build first, learn as needed"

#### Accountability Avoidance Detection

**Negative (Medium):**
- Keywords: "just for me", "personal project"
- Message: "Solo-only project - no external accountability"
- Suggestion: "Add public component or external deadline"

**Positive:**
- Keywords: "public", "share", "github"
- Message: "External accountability component detected"

---

## Database Schema

### Tables

#### ideas

```sql
CREATE TABLE IF NOT EXISTS ideas (
    id TEXT PRIMARY KEY,
    content TEXT NOT NULL,
    raw_score REAL,
    final_score REAL,
    patterns TEXT,                  -- JSON-serialized Vec<String>
    recommendation TEXT,
    analysis_details TEXT,          -- JSON-serialized Score struct
    created_at TEXT NOT NULL,       -- RFC3339 format
    reviewed_at TEXT,               -- RFC3339 format
    status TEXT NOT NULL DEFAULT 'active'
);
```

**Indexes:**
```sql
CREATE INDEX IF NOT EXISTS idx_ideas_created_at ON ideas(created_at);
CREATE INDEX IF NOT EXISTS idx_ideas_final_score ON ideas(final_score);
CREATE INDEX IF NOT EXISTS idx_ideas_status ON ideas(status);
CREATE INDEX IF NOT EXISTS idx_ideas_status_score ON ideas(status, final_score);
```

#### idea_relationships

```sql
CREATE TABLE IF NOT EXISTS idea_relationships (
    id TEXT PRIMARY KEY,
    source_idea_id TEXT NOT NULL,
    target_idea_id TEXT NOT NULL,
    relationship_type TEXT NOT NULL,
    created_at TEXT NOT NULL,       -- RFC3339 format
    FOREIGN KEY (source_idea_id) REFERENCES ideas (id),
    FOREIGN KEY (target_idea_id) REFERENCES ideas (id)
);
```

**Indexes:**
```sql
CREATE INDEX IF NOT EXISTS idx_relationships_source ON idea_relationships(source_idea_id);
CREATE INDEX IF NOT EXISTS idx_relationships_target ON idea_relationships(target_idea_id);
CREATE INDEX IF NOT EXISTS idx_relationships_type ON idea_relationships(relationship_type);
```

### Connection Pool Configuration

```rust
SqlitePoolOptions::new()
    .max_connections(2)
    .min_connections(1)
    .idle_timeout(Duration::from_secs(5 * 60))  // 5 minutes
    .acquire_timeout(Duration::from_secs(5))
```

### Retry Configuration

```rust
pub struct RetryConfig {
    pub max_attempts: u32,           // Default: 3
    pub base_delay_ms: u64,          // Default: 100
    pub max_delay_ms: u64,           // Default: 5000
    pub backoff_multiplier: f64,     // Default: 2.0
    pub jitter: bool,                // Default: true
    pub retryable_errors: Vec<String>,
}
```

**Default Retryable Errors:**
- "database is locked"
- "database table is locked"
- "connection refused"
- "connection timeout"
- "connection reset"
- "temporary failure"
- "busy"
- "timeout"

---

## CLI Commands

### tm dump [IDEA]

**Purpose:** Quick-capture an idea with analysis

**Flags:**
- `--interactive`, `-i`: Run in interactive loop mode
- `--quick`, `-q`: Save without analysis
- `--no-ai`: Use rule-based scoring (no LLM)
- `--force-claude`: Force use of Claude CLI instead of Ollama

**Behavior:**
1. Accept idea text from:
   - Command argument
   - Clipboard (automatic detection)
   - Multi-line input (Ctrl+D to finish)
2. If `--quick`: Save immediately without scoring
3. Otherwise: Analyze using scoring engine
4. Save to database with:
   - Generated UUID
   - Score breakdown
   - Pattern matches
   - Recommendation
   - Analysis details (JSON)
   - Timestamp (UTC, RFC3339)
5. Display results with emoji formatting
6. Prompt for next actions

**Output Format:**
```
🎯 Final Score: 8.5/10
🔥 PRIORITIZE NOW

📊 Breakdown:
  Mission Alignment: 3.20/4.00
  Anti-Challenge: 2.80/3.50
  Strategic Fit: 2.00/2.50

⚠️ Patterns Detected:
  🔄 Context-Switching: Staying focused on current tech stack

✅ Idea saved (ID: abc123...)
```

### tm analyze [IDEA]

**Purpose:** Analyze an idea without saving

**Flags:**
- `--last`: Analyze the most recently saved idea
- `--no-ai`: Use rule-based analysis

**Behavior:**
1. If `--last`: Fetch last idea from database
2. Otherwise: Analyze provided idea text
3. Calculate score and detect patterns
4. Display detailed analysis (same format as dump)
5. **Do NOT save to database**

### tm review

**Purpose:** Review recently saved ideas

**Flags:**
- `--limit N`: Number of ideas to show (default: 10)
- `--min-score X`: Filter by minimum score (default: 0.0)

**Behavior:**
1. Query database: `ORDER BY created_at DESC LIMIT N`
2. Filter by `min-score` and `status = 'active'`
3. Display list with:
   - ID (first 8 chars)
   - Score
   - Recommendation emoji
   - Content preview (first 60 chars)
   - Created timestamp (relative: "2 hours ago")

**Output Format:**
```
📋 Recent Ideas (10 most recent):

1. abc12345 | 8.5 🔥 | "Build an AI automation tool..." | 2 hours ago
2. def67890 | 6.2 ✅ | "Create a Python script for..." | 1 day ago
3. ghi11111 | 3.1 🚫 | "Learn Rust before building..." | 3 days ago
```

### Database Operations

#### save_idea()

```rust
async fn save_idea(
    &self,
    content: &str,
    raw_score: Option<f64>,
    final_score: Option<f64>,
    patterns: Option<Vec<String>>,
    recommendation: Option<String>,
    analysis_details: Option<String>,
) -> DatabaseResult<String>
```

**Returns:** UUID string

**Timeout:** 30 seconds

**Retry:** Up to 3 attempts with exponential backoff

#### get_last_idea()

```rust
async fn get_last_idea(&self) -> DatabaseResult<Option<StoredIdea>>
```

**Query:**
```sql
SELECT * FROM ideas
WHERE status = 'active'
ORDER BY created_at DESC
LIMIT 1
```

**Timeout:** 10 seconds

#### get_ideas_with_filters()

```rust
async fn get_ideas_with_filters(
    &self,
    limit: usize,
    min_score: f64,
) -> DatabaseResult<Vec<StoredIdea>>
```

**Query:**
```sql
SELECT * FROM ideas
WHERE status = 'active' AND (final_score >= ?1 OR final_score IS NULL)
ORDER BY created_at DESC
LIMIT ?2
```

---

## Test Cases

### High Score Example

**Input:**
```
Build an AI automation tool using Python and LangChain to help
hotel staff automate guest request routing. Can ship MVP in 30 days.
Target $2K/month recurring revenue. Will build in public on Twitter.
```

**Expected Score:** ~8.5-9.0

**Breakdown:**
- **Mission:**
  - Domain Expertise: 1.1 (uses Python, hotel domain)
  - AI Alignment: 1.4 (core AI product)
  - Execution Support: 0.75 (30-day MVP)
  - Revenue Potential: 0.45 ($2K/month)
  - **Total: ~3.7/4.0**

- **Anti-Challenge:**
  - Context Switching: 1.15 (uses current stack)
  - Rapid Prototyping: 0.95 (30-day MVP)
  - Accountability: 0.7 (public building)
  - Income Anxiety: 0.45 (fast revenue)
  - **Total: ~3.25/3.5**

- **Strategic:**
  - Stack Compatibility: 0.9 (Python flow sessions)
  - Shipping Habit: 0.7 (reusable AI components)
  - Public Accountability: 0.35 (Twitter validation)
  - Revenue Testing: 0.28 (SaaS model)
  - **Total: ~2.23/2.5**

**Final Score:** ~9.2/10
**Recommendation:** 🔥 PRIORITIZE NOW

### Low Score Example

**Input:**
```
Learn Rust and build a comprehensive game engine from scratch.
Will need 6 months to learn the basics first, then another 6 months
to build a production-ready system. Personal project for fun.
```

**Expected Score:** ~2.0-3.0

**Breakdown:**
- **Mission:**
  - Domain Expertise: 0.1 (no matching skills)
  - AI Alignment: 0.0 (no AI component)
  - Execution Support: 0.05 (learning-focused)
  - Revenue Potential: 0.0 (no revenue path)
  - **Total: ~0.15/4.0**

- **Anti-Challenge:**
  - Context Switching: 0.1 (complete stack switch - Rust)
  - Rapid Prototyping: 0.1 (perfection-dependent)
  - Accountability: 0.05 (personal project)
  - Income Anxiety: 0.0 (no revenue)
  - **Total: ~0.25/3.5**

- **Strategic:**
  - Stack Compatibility: 0.15 (unclear execution)
  - Shipping Habit: 0.65 (reusable engine)
  - Public Accountability: 0.05 (6+ months)
  - Revenue Testing: 0.0 (no revenue model)
  - **Total: ~0.85/2.5**

**Final Score:** ~1.3/10
**Recommendation:** 🚫 AVOID FOR NOW

**Patterns Detected:**
- 🔄 Context-Switching (High): New tech stack (Rust)
- 🕰️ Procrastination (Critical): Learning before building
- 👤 Accountability Avoidance (Medium): Personal project only

### Medium Score Example

**Input:**
```
Create a Python script to automate my daily standup notes.
Will use it personally to save 15 minutes per day. Should take
about 2 weeks to build a working version.
```

**Expected Score:** ~5.5-6.5

**Breakdown:**
- **Mission:**
  - Domain Expertise: 1.0 (Python)
  - AI Alignment: 0.1 (automation but not AI-focused)
  - Execution Support: 0.7 (2-week timeline)
  - Revenue Potential: 0.05 (personal use)
  - **Total: ~1.85/4.0**

- **Anti-Challenge:**
  - Context Switching: 1.1 (current stack)
  - Rapid Prototyping: 0.9 (2 weeks)
  - Accountability: 0.1 (personal use)
  - Income Anxiety: 0.0 (no revenue)
  - **Total: ~2.1/3.5**

- **Strategic:**
  - Stack Compatibility: 0.8 (Python flow)
  - Shipping Habit: 0.6 (reusable script)
  - Public Accountability: 0.3 (quick validation)
  - Revenue Testing: 0.05 (no revenue model)
  - **Total: ~1.75/2.5**

**Final Score:** ~5.7/10
**Recommendation:** ⚠️ CONSIDER LATER

---

## Implementation Notes for Go

### Type System Differences

1. **Newtype Pattern:** Go doesn't have Rust's newtype pattern. Use struct wrappers:
   ```go
   type IdeaID struct {
       value string
   }
   ```

2. **Option Types:** Use pointers or explicit `Valid` booleans:
   ```go
   type Score struct {
       Value float64
       Valid bool
   }
   ```

3. **Result Types:** Use explicit error returns:
   ```go
   func Calculate(idea string) (*Score, error)
   ```

### Concurrency

Rust version uses `tokio::join!` for concurrent scoring. In Go, use goroutines:
```go
var score Score
var patterns []Pattern
var wg sync.WaitGroup

wg.Add(2)
go func() {
    defer wg.Done()
    score = engine.Calculate(idea)
}()
go func() {
    defer wg.Done()
    patterns = detector.Detect(idea)
}()
wg.Wait()
```

### Database

- Use `database/sql` with SQLite driver
- Implement connection pool with similar settings
- Implement retry logic with exponential backoff
- Use `time.RFC3339` for timestamps (matches Rust's RFC3339)

### JSON Serialization

- Use `encoding/json` for patterns and analysis_details
- Match Rust's serde format exactly

### Testing

- Create table-driven tests for each scoring dimension
- Test all regex patterns
- Test database operations with retry logic
- Test Telos parser with example file

---

## End of Reference

This document should be updated as the Rust implementation evolves. When implementing the Go version, maintain **exact** behavioral parity with this specification.
