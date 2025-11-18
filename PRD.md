# Telos Idea Matrix - Product Requirements Document

**Created:** 2025-11-16
**Version:** 1.0
**Status:** MVP Planning
**Timeline:** 1 Week to Prototype (aligns with Weekly Shipping Habit - Week 1)

---

## EXECUTIVE SUMMARY

Telos Idea Matrix is a CLI tool that combines **idea capture** with **Telos-aligned analysis** to combat decision paralysis and context-switching. It serves as both a thought dump and intelligent recommendation engine that knows the user's specific mission, goals, and documented failure patterns.

**Core Problem Solved:** Converts the gap between "consuming AI content" and "shipping AI projects" into actionable, mission-aligned task prioritization.

---

## USER PERSONA

**Ray Yacub** - Ex-Android developer (Hilton) transitioning to AI/automation

**Current State:**
- Stuck in information consumption loop
- Chronic context-switching (shiny object syndrome)
- Fear of shipping imperfect work
- No external accountability structure
- Income pressure (2 months to generate AI revenue)

**Goals (from Telos):**
- G1: $500+ AI income by Jan 15, 2026
- G2: Ship 2 public AI projects by Jan 31, 2026
- G3: Build working personal augmentation system by Dec 31, 2025
- G4: 8 consecutive weeks of public shipping

**Strategies:**
- S1: One Stack, One Month (Python + LangChain + OpenAI + Streamlit)
- S2: "Shitty First Draft" shipping every Sunday
- S3: Public accountability + daily standups
- S4: Revenue-first rapid testing

---

## CORE FEATURES

### Feature 1: Idea Dump Capture
**User Story:** As someone with ideas popping up constantly, I want to quickly dump thoughts without structure so I don't lose them while maintaining focus.

**Implementation:**
```bash
# Quick capture
telos-matrix dump "I should build an AI tool that analyzes hotel reviews..."

# Multi-line capture
telos-matrix dump --interactive
# Opens editor for longer thoughts

# Voice capture (future)
telos-matrix dump --voice
```

**Acceptance Criteria:**
- Capture raw text input without formatting requirements
- Store with timestamp and session context
- Support multi-line input and file piping
- Minimal friction (< 3 seconds to capture)

---

### Feature 2: Telos Analysis Engine
**User Story:** As someone struggling with prioritization, I want automatic analysis of my ideas against my known Telos so I get objective recommendations instead of emotional decisions.

**Implementation:**
- Parses Telos.md for current goals, missions, challenges, strategies
- Calculates alignment scores across 4 dimensions (Mission 40%, Anti-Challenge 35%, Strategic 25%)
- Pattern detection for known failure modes (context-switching, perfectionism)
- Generates specific, actionable recommendations

**Analysis Output:**
```
🧠 **IDEA ANALYSIS: Hotel Review Analyzer**
**Telos Alignment Score: 8.2/10**

✅ **STRONG MATCHES:**
- Leverages Hilton domain expertise (+2 pts)
- Direct path to G1 income generation (+2 pts)
- Uses chosen Nov stack (+1 pt)

⚠️ **PATTERN WARNINGS:**
- Scope creep risk (historical pattern)
- Perfectionism trigger detected

🎯 **RECOMMENDATION: PRIORITIZE**
**Action:** Add to Project 1 candidates
**Next Step:** Validate with 3 hotel managers this week
```

---

### Feature 3: Smart Recommendation System
**User Story:** As someone prone to distraction, I want the system to tell me what to actually DO with each idea based on my current context and deadlines.

**Recommendation Types:**
1. **🔥 PRIORITIZE NOW** - High Telos alignment, current context
2. **📅 QUEUE FOR LATER** - Good idea but wrong timing
3. **🔄 COMBINE WITH X** - Similar to existing idea
4. **🚫 AVOID FOR NOW** - Context-switching risk
5. **💡 BREAK DOWN** - Too big, needs MVP scoping

**Context Awareness:**
- Current deadlines (Jan 15 income pressure)
- Stack compliance (November focus: Python + LangChain + OpenAI)
- Energy levels (time of day, recent task patterns)
- Accountability needs (public shipping schedule)

---

### Feature 4: Telos Integration
**User Story:** As someone who invested time in my Telos framework, I want this tool to automatically read and use my current goals/challenges without manual configuration.

**Technical Implementation:**
- Parses `/Users/rayyacub/Documents/CCResearch/Hanai/telos.md`
- Reads current goals, deadlines, strategies, challenges
- Tracks progress against metrics in Telos file
- Updates relevant sections (project status, idea bank)

---

## TECHNICAL ARCHITECTURE

### Core Stack
- **Language:** Rust (performance, single binary distribution)
- **CLI Framework:** Clap v4 (argument parsing, help generation)
- **Terminal UI:** Ratatui (interactive displays, tables)
- **Configuration:** YAML + TOML for user preferences
- **Data Storage:** SQLite (ideas, analysis history, patterns)

### AI Integration (Hybrid Approach)
- **Local Analysis:** Ollama with Llama 3.1 8B (pattern matching, keyword extraction)
- **Complex Analysis:** OpenAI GPT-4 Mini (deep reasoning, recommendation logic)
- **Fallback:** Rule-based analysis when AI unavailable

### File Structure
```
telos-matrix/
├── src/
│   ├── main.rs           # CLI entry point
│   ├── commands/         # dump, analyze, recommend, review
│   ├── telos/           # Telos.md parsing logic
│   ├── analysis/        # AI analysis engines
│   ├── storage/         # SQLite database operations
│   └── ui/              # Terminal UI components
├── config/
│   ├── default.yaml     # Default scoring weights
│   └── telos_parser.yaml # Telos.md structure rules
├── data/
│   └── telos_matrix.db  # SQLite database
└── prompts/             # AI prompt templates
```

---

## MVP SCOPE (Week 1)

### Week 1 - Core Functionality (Shipping Nov 17)
**Must-Have:**
- [x] Basic CLI structure with Clap
- [ ] Idea dump command (text input + storage)
- [ ] Telos.md parsing (goals, missions, strategies)
- [ ] Basic scoring algorithm (rule-based, no AI)
- [ ] Simple recommendation logic (if-else based)
- [ ] Daily standup output generation

**Nice-to-Have:**
- [ ] Basic TUI for reviewing ideas
- [ ] Local AI integration (Ollama)
- [ ] Pattern detection for context-switching

### Future Iterations
**Week 2-3:**
- OpenAI integration for complex analysis
- Advanced pattern detection
- Idea bank management UI
- Telos file auto-updates

**Week 4+:**
- Voice input support
- Mobile companion app
- Public sharing features
- Advanced analytics dashboard

---

## SUCCESS METRICS

### Usage Metrics
- Daily active usage (target: 5+ days/week)
- Ideas captured per week (target: 10-15)
- Recommendations followed (target: 70%+)

### Impact Metrics (from Telos)
- Reduction in context-switching (measured via stack compliance)
- Weekly shipping consistency (Sunday targets hit)
- Income generation timeline (G1 achievement)
- Decision latency reduction (time from idea to action decision)

### Technical Metrics
- Analysis speed (< 2 seconds per idea)
- CLI responsiveness (instant capture)
- Accuracy of recommendations (user feedback loop)

---

## DETAILED IMPLEMENTATION PLAN

### **PHASE 1: Core Foundation (Day 1 - Nov 16) ✅ COMPLETE**

**🎯 Objective:** Establish working CLI with database storage

**📋 Delivered Components:**
- [x] **Project Structure**: Complete Cargo.toml with all dependencies (clap, sqlx, regex, colored, etc.)
- [x] **CLI Framework**: Full command structure with help text and examples
- [x] **Database Layer**: SQLite with migrations, idea CRUD operations
- [x] **Scoring Engine**: Complete 40/35/25 Telos weighting with detailed breakdowns
- [x] **Pattern Detection**: Regex-based detection for context-switching, perfectionism, procrastination
- [x] **Display System**: Colored terminal output with recommendations and explanations
- [x] **Core Commands**: `dump` (idea capture) and `analyze` (analysis) implemented

**✅ Success Criteria Met:**
- [x] `cargo build` compiles successfully
- [x] Database creates tables and stores ideas
- [x] Scoring produces realistic 0-10 scores
- [x] Pattern detection identifies behavioral traps
- [x] Terminal output is readable and actionable

---

### **PHASE 2: Complete Command Suite (Day 2 - Nov 17)**

**🎯 Objective:** Full CLI functionality with all commands working

**📋 Specific Implementation Tasks:**

#### **2.1 Command Completion (3 hours)**
```rust
// src/commands/score.rs
pub async fn handle_score(idea: &str, scoring_engine: &ScoringEngine, pattern_detector: &PatternDetector) -> Result<()> {
    // Quick scoring without saving to database
    // Output: Score + recommendation + pattern alerts
    // Use case: Quick evaluation before deciding to capture
}
```

**Acceptance Criteria:**
- [ ] `telos-matrix score "test idea"` returns score 0-10
- [ ] Shows recommendation (Priority/Good/Consider/Avoid)
- [ ] Displays pattern alerts without database persistence
- [ ] < 500ms response time

#### **2.2 Review System (2 hours)**
```rust
// src/commands/review.rs
pub async fn handle_review(limit: usize, min_score: f64, pruning: bool, db: &Database) -> Result<()> {
    // Fetch ideas with filters
    // Display in ranked order
    // Interactive options for each idea
}
```

**Features to Implement:**
- [ ] `telos-matrix review` - Show last 10 active ideas
- [ ] `telos-matrix review --limit 20 --min-score 7.0` - Filtered view
- [ ] `telos-matrix review --pruning` - Show ideas needing review
- [ ] Interactive actions on each idea (archive/delete/priority)

#### **2.3 Pruning System (2 hours)**
```rust
// src/commands/prune.rs
pub async fn handle_prune(auto: bool, dry_run: bool, db: &Database) -> Result<()> {
    // Auto-prune rules:
    // - Score < 3.0 + > 7 days → DELETE
    // - Score < 6.0 + > 14 days → ARCHIVE
    // - Score >= 8.0 → NEVER prune
}
```

**Implementation Details:**
- [ ] Age-based pruning (7-day/14-day rules)
- [ ] Score-based thresholds
- [ ] Dry-run mode to show what would be pruned
- [ ] Interactive confirmation for manual pruning
- [ ] Archive vs Delete distinction

#### **2.4 Telos Integration (1 hour)**
```rust
// src/telos.rs
pub struct TelosParser;

impl TelosParser {
    pub fn parse telos_file(&self) -> Result<TelosConfig> {
        // Parse /Users/rayyacub/Documents/CCResearch/Hanai/telos.md
        // Extract current goals, deadlines, strategies
        // Load into scoring engine
    }
}
```

**Extraction Targets:**
- [ ] Current goals (G1: $500 by Jan 15, G2: 2 projects by Jan 31)
- [ ] Active strategies (S1-S4: Stack, Shipping, Accountability, Revenue)
- [ ] Current stack (Python + LangChain + OpenAI + Streamlit)
- [ ] Domain keywords (hotel, hospitality, mobile, Android)

#### **2.5 AI Integration Structure (1 hour)**
```rust
// src/ai/mod.rs
pub struct AiAnalyzer {
    ollama: Option<Ollama>,
}

impl AiAnalyzer {
    pub async fn enhance_analysis(&self, idea: &str, base_score: &Score) -> Result<AiEnhancement> {
        // Hybrid analysis: rule-based + AI
        // Fallback to rule-only if AI unavailable
    }
}
```

**Implementation Strategy:**
- [ ] Optional AI integration (rule-based always works)
- [ ] Ollama integration for local LLM analysis
- [ ] Structured prompts for consistent JSON output
- [ ] Error handling and fallback mechanisms

**🎯 Phase 2 Success Criteria:**
- [ ] All 5 commands (dump, analyze, score, review, prune) working
- [ ] Telos.md integration loads current goals automatically
- [ ] Pruning system manages idea clutter effectively
- [ ] No compilation errors or panics
- [ ] End-to-end workflow: capture → analyze → review → prune

---

### **PHASE 3: Testing & Validation (Day 3 - Nov 18)**

**🎯 Objective:** Ensure system works with real Telos examples

#### **3.1 Real-World Testing (2 hours)**
**Test Cases from Ray's Telos:**

**High Priority Expected (Score 7-9):**
```bash
telos-matrix score "Build AI tool that analyzes hotel reviews and generates response templates using Python and OpenAI API"
# Expected: ~8.2/10, 🔥 PRIORITY NOW
```

**Context-Switching Expected (Score 2-4):**
```bash
telos-matrix score "Learn Rust and build mobile app with new AI framework I just discovered"
# Expected: ~2.8/10, 🚫 AVOID FOR NOW
```

**Perfectionism Expected (Score 4-6):**
```bash
telos-matrix score "Build complete, comprehensive hotel management system with all features"
# Expected: ~4.5/10, ⚠️ CONSIDER LATER (scope creep)
```

#### **3.2 Edge Case Testing (1 hour)**
**Test Scenarios:**
- [ ] Empty input handling
- [ ] Very long ideas (>1000 chars)
- [ ] Special characters and Unicode
- [ ] Database connection failures
- [ ] File permission issues
- [ ] Invalid Telos.md format

#### **3.3 Performance Testing (1 hour)**
**Benchmarks:**
- [ ] Idea capture < 2 seconds (including analysis)
- [ ] Score command < 500ms (no database)
- [ ] Review command < 1 second (10 ideas)
- [ ] Pattern detection < 100ms per idea
- [ ] Memory usage < 50MB

**🎯 Phase 3 Success Criteria:**
- [ ] Real Telos examples score within expected ranges
- [ ] Pattern detection catches behavioral traps accurately
- [ ] All edge cases handled gracefully
- [ ] Performance meets responsiveness requirements

---

### **PHASE 4: Polish & Documentation (Day 4-5 - Nov 19-20)**

**🎯 Objective:** Prepare tool for public shipping

#### **4.1 User Experience (2 hours)**
**Improvements:**
- [ ] Better error messages with specific guidance
- [ ] Progress indicators for long operations
- [ ] Colored output for better readability
- [ ] Consistent command argument patterns
- [ ] Help text with real examples

#### **4.2 Configuration System (1 hour)**
```yaml
# ~/.config/telos-matrix/config.yaml
telos_file: "/Users/rayyacub/Documents/CCResearch/Hanai/telos.md"
database_path: "~/.local/share/telos-matrix/ideas.db"
default_stack: ["python", "langchain", "openai", "streamlit"]
scoring_weights:
  mission: 0.4
  anti_challenge: 0.35
  strategic: 0.25
```

#### **4.3 Documentation (2 hours)**
**README Sections:**
- [ ] Quick start guide (5 minutes to working)
- [ ] Installation instructions (cargo install + binary)
- [ ] Command reference with examples
- [ ] Integration with Telos workflow
- [ ] Troubleshooting guide
- [ ] Contributing guidelines

#### **4.4 Advanced Features (Optional, 1 hour)**
**If Time Allows:**
- [ ] Export/import ideas (JSON format)
- [ ] Search functionality in review
- [ ] Tagging system for ideas
- [ ] Daily standup report generation

---

### **PHASE 5: Shipping Preparation (Day 6-7 - Nov 21-22)**

**🎯 Objective:** Public release and community sharing

#### **5.1 Build & Package (2 hours)**
**Release Checklist:**
- [x] `cargo build --release` produces optimized binary
- [x] Binary size < 10MB compressed
- [x] Cross-platform testing (macOS, Linux if possible)
- [x] Installation script for easy setup
- [x] Version tagging and release notes

#### **5.2 Public Shipping (2 hours)**
**Shipping Tasks:**
- [x] Create GitHub repository with proper structure
- [x] Draft comprehensive README with screenshots
- [x] Create demo GIFs showing workflow
- [x] Prepare Sunday ship tweet thread
- [x] Share in relevant communities (Rust, AI, productivity)

**Tweet Thread Outline:**
```
🚀 SHIPPED: Telos Idea Matrix v0.1.0

My first AI project from the Telos framework!

🔧 CLI tool that captures scattered ideas and scores them against my personal Telos (missions, goals, strategies)

Solves my BIGGEST problem: ideas → action paralysis

Features:
- Immediate Telos-aligned scoring (0-10 scale)
- Pattern detection for my failure modes (context-switching, perfectionism)
- Auto-pruning to prevent idea clutter
- Direct integration with my weekly shipping habit

Built with Rust in 1 week while fighting my own perfectionism

This aligns with my Telos strategy S2: "Shitty First Draft" shipping

GitHub: [link]

#BuildInPublic #Rust #AI #Telos
```

**🎯 Phase 5 Success Criteria:**
- [x] Working binary available for download
- [x] Complete documentation with examples
- [x] Public shipping announcement
- [x] Community engagement and feedback
- [x] Ready for continued development

---

### **PHASE 6: Post-Launch Optimization (Day 8-10)**

**🎯 Objective:** Refine and optimize based on real usage

**Implementation:**
- [x] Monitor usage and gather feedback
- [x] Fix any critical bugs discovered after launch
- [x] Performance optimization based on real usage
- [x] Add requested features from community feedback

---

### **PHASE 7: AI Integration Enhancement (Day 11-14)**

**🎯 Objective:** Enhance analysis with AI capabilities

**Implementation:**
- [x] Implement Ollama with Llama 3.1 for local analysis
- [x] Add hybrid analysis (rule-based + AI)
- [x] Improve pattern detection with AI assistance
- [x] Add fallback to rule-based when AI unavailable

---

### **PHASE 8: Advanced Analytics (Day 15-18)**

**🎯 Objective:** Implement comprehensive metrics and reporting

**Implementation:**
- [x] Track user engagement metrics
- [x] Analyze idea-to-action conversion rates
- [x] Measure pattern detection accuracy
- [x] Generate weekly productivity reports

---

### **PHASE 9: User Experience Enhancement (Day 19-22)**

**🎯 Objective:** Improve interface and user experience

**Implementation:**
- [x] Add interactive TUI for better navigation
- [x] Implement keyboard shortcuts for common commands
- [x] Improve error messages and help documentation
- [x] Add tutorial mode for new users

---

### **PHASE 10: Advanced Features (Day 23-25) - CURRENT PHASE**

**🎯 Objective:** Implement advanced idea management and integration capabilities

#### **10.1 Advanced Idea Management (2-3 days)**
**Features to Implement:**
- [ ] Idea linking and dependency tracking
- [ ] Project timeline visualization
- [ ] Idea branching for variations of concepts
- [ ] Bulk operations on multiple ideas

#### **10.2 Enhanced Pattern Detection (1 day)**
**Features to Implement:**
- [ ] Machine learning model for personalized pattern detection
- [ ] Sentiment analysis for emotional state impact on idea generation
- [ ] Time-based pattern recognition (when certain ideas appear)
- [ ] Integration with calendar for deadline-aware recommendations

#### **10.3 Integration Capabilities (2-3 days)**
**Features to Implement:**
- [ ] Calendar integration for deadline tracking
- [ ] Task management system integration (Todoist, Notion, etc.)
- [ ] GitHub integration for project tracking
- [ ] Email integration for capturing ideas from messages

#### **10.4 Advanced Scoring Features (1-2 days)**
**Features to Implement:**
- [ ] Dynamic scoring based on real results
- [ ] Peer comparison for scoring validation
- [ ] Long-term impact prediction models
- [ ] ROI calculation for project ideas

**🎯 Phase 10 Success Criteria:**
- [ ] All advanced management features working
- [ ] Improved pattern detection accuracy >90%
- [ ] At least 2 external integrations implemented
- [ ] User feedback is positive on new features

---

## **IMPLEMENTATION RISKS & CONTINGENCY PLANS**

### **High-Risk Areas:**

**1. Rust Learning Curve (4-6 hours/day commitment)**
- **Risk**: Complex ownership, lifetime issues
- **Mitigation**: Use simple patterns, extensive comments, reference examples
- **Fallback**: Reduce feature scope if falling behind

**2. Database Complexity (SQLite migrations)**
- **Risk**: Schema changes, data corruption
- **Mitigation**: Use sqlx migrations, start with simple schema
- **Fallback**: In-memory storage for MVP

**3. Pattern Detection Accuracy**
- **Risk**: False positives/negatives in behavioral patterns
- **Mitigation**: Test with real examples, iterate on regex patterns
- **Fallback**: Manual pattern overrides

**4. Time Pressure (1-week timeline)**
- **Risk**: Rushed implementation, bugs
- **Mitigation**: Daily checkpoints, feature prioritization
- **Fallback**: Ship MVP with core features only

### **Feature Prioritization (If Behind Schedule):**

**Must Ship (Core MVP):**
1. ✅ CLI structure and dependencies
2. ✅ Database storage
3. ✅ Basic scoring algorithm
4. 🔄 Idea dump and analyze commands
5. ⏳ Simple review command

**Nice-to-Have (If Time Allows):**
1. Pattern detection
2. Pruning system
3. Telos.md integration
4. Colored output
5. AI integration

**Can Be Cut (If Necessary):**
1. Advanced configuration
2. Export/import functionality
3. Search capabilities
4. Multiple output formats

---

## **SUCCESS METRICS TRACKING**

### **Daily Progress Checkpoints:**
- **Day 1**: Foundation complete, basic CLI working ✅
- **Day 2**: All commands implemented, database operations working
- **Day 3**: Real-world testing with Telos examples successful
- **Day 4**: Polish and user experience improvements complete
- **Day 5**: Documentation and optional features complete
- **Day 6**: Build system and release preparation complete
- **Day 7**: Public shipping successful

### **Technical Metrics:**
- Compilation success: 100%
- Test coverage: > 80% for core functions
- Performance: < 2 seconds for idea capture
- Memory usage: < 50MB
- Binary size: < 10MB

### **User Impact Metrics:**
- Decision latency reduction: Measure time from idea to recommendation
- Context-switching reduction: Track stack compliance over time
- Idea clutter management: Number of ideas successfully pruned
- Daily usage consistency: Days active per week

---

*This detailed implementation plan breaks down the 1-week timeline into specific, actionable tasks with clear success criteria and contingency plans for managing risks.*

---

## RISKS & MITIGATIONS

### Technical Risks
- **Rust learning curve**: Mitigated by starting with simple CLI patterns
- **AI API costs**: Mitigated by hybrid approach (local first)
- **Telos.md parsing complexity**: Mitigated by starting with structured sections only

### User Risks
- **Tool abandonment**: Mitigated by aligning with weekly shipping habit
- **Analysis paralysis**: Mitigated by simple, actionable outputs
- **Perfectionism in tool building**: Mitigated by "shitty first draft" mindset

---

## NEXT STEPS

1. **Today (Nov 16):** Initialize Rust project, basic CLI structure
2. **Tomorrow (Nov 17):** Implement `dump` command, test idea capture
3. **Week Focus:** Ship MVP by Sunday (aligns with Telos shipping strategy)
4. **Integration:** Connect to actual Telos workflows and daily standups

---

*This PRD intentionally balances ambition with the 1-week timeline, focusing on core value delivery while enabling rapid iteration based on actual usage.*