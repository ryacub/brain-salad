# Alex's 2025 Engineering Telos

**Persona**: Alex Chen, Senior Software Engineer (5 YoE)
**Current Role**: Senior Engineer at mid-size SaaS company
**Career Target**: Staff/Principal Engineer by end of 2026
**Location**: Remote, Pacific timezone
**Last Updated**: 2025-01-15

## About This Telos

I've been a software engineer for 5 years, and I've hit a plateau. I'm technically competent, but I'm not standing out. I see other engineers at my level getting promoted, speaking at conferences, and building reputations in the industry while I'm stuck in the day-to-day grind.

My problem isn't lack of ideas or motivation. It's the opposite. I start too many things, jump between technologies, and rarely finish what I start. I watch tutorials instead of building. I plan blog posts instead of publishing them. I need to focus, ship, and build credibility in ONE ecosystem.

This Telos is my commitment to depth over breadth, shipping over perfecting, and building a reputation in the Rust ecosystem.

---

## Goals

### G1: Become a recognized Rust ecosystem contributor (Deadline: 2025-12-31)
- **Target**: Merge 50+ meaningful PRs to established Rust projects
- **Key Projects**: Tokio, Serde, Axum, SQLx, or similar (pick 2-3 max)
- **Success Criteria**:
  - GitHub profile shows sustained contributions
  - Maintainers recognize my username
  - At least 1 significant feature or major bug fix
- **Why**: Staff engineers are known beyond their company. Contributing to widely-used projects builds credibility and demonstrates deep expertise.

### G2: Publish 12 technical blog posts in 2025 (Deadline: 2025-12-31)
- **Cadence**: One post per month minimum
- **Topics**: Rust internals, systems programming, performance optimization, distributed systems
- **Platform**: Personal blog (Astro + Markdown, hosted on Cloudflare Pages)
- **Success Criteria**:
  - At least 3 posts get 100+ upvotes on Reddit/HN
  - At least 1 post is shared by Rust community influencers
  - Blog averages 500+ monthly visitors by December
- **Why**: Writing clarifies thinking, demonstrates expertise, and builds a personal brand. Staff engineers are teachers.

### G3: Speak at 2 technical conferences (Deadline: 2025-12-31)
- **Q1-Q2**: Submit talks to RustConf, RustNL, Oxidize
- **Q3-Q4**: Local meetups as practice (SF Rust Meetup, Seattle Rust)
- **Talk Topics**:
  - "Building Production-Grade Async Services in Rust"
  - "Zero-Copy Patterns in Rust: Performance Without Compromise"
  - "Debugging Async Rust: Beyond println!"
- **Success Criteria**:
  - At least 1 accepted talk at a major conference
  - At least 1 local meetup presentation
  - Record and publish talks on YouTube
- **Why**: Public speaking amplifies reach and credibility. Staff engineers represent their company and the broader community.

### G4: Build and ship 2 production-quality side projects (Deadline: 2025-12-31)
- **Project 1 (Q1-Q2)**: Rust CLI tool solving a real problem (NOT a toy)
  - Example: Advanced database migration tool, log aggregation utility, or performance profiler
  - Success: 100+ GitHub stars, 10+ users giving feedback
- **Project 2 (Q3-Q4)**: Backend service or library
  - Example: Message queue, caching layer, or API rate limiter
  - Success: Used in production somewhere (even if just my own projects)
- **Why**: Side projects demonstrate initiative, technical breadth, and ability to ship. They're portfolio pieces for promotion conversations.

### G5: Get promoted to Staff Engineer (Deadline: 2026-06-30)
- **2025 Focus**: Build external credibility and deep technical expertise
- **Internal Actions**:
  - Lead 2 cross-team technical initiatives
  - Mentor 2 junior engineers
  - Write design docs for major features
  - Own architecture decisions for team's services
- **External Actions**: Everything in G1-G4
- **Success Criteria**:
  - Promotion packet approved by end of Q2 2026
  - External credibility (blog, talks, OSS) cited as evidence
- **Why**: This is the ultimate goal. Everything else supports this.

---

## Strategies

### S1: Deep Work Sessions (4+ hours, 3x per week minimum)
- **When**: Early mornings before work (6am-10am) or weekends
- **What**: Uninterrupted coding, writing, or learning
- **Where**: Home office, coffee shop, or library (NO open office)
- **Rules**:
  - Phone on airplane mode
  - No Slack, email, or social media
  - One task only (no context switching)
  - Track hours using Toggl
- **Why**: Complex technical work requires sustained focus. Staff-level work cannot be done in 30-minute chunks between meetings.

### S2: Learn in Public
- **Every project generates content**:
  - Working on open source? Write about what I learned
  - Built a feature? Extract lessons and blog about it
  - Solved a tough bug? Document the debugging process
- **Post everywhere**:
  - Personal blog (primary)
  - Reddit (r/rust, r/programming)
  - Hacker News
  - Dev.to / Hashnode (cross-post)
  - Twitter/X (share snippets and insights)
- **Why**: Public learning creates accountability, builds audience, and documents growth. It's also how you get discovered.

### S3: One Language Rule
- **Primary Language**: Rust (90% of effort)
- **Secondary Language**: Go (only for work requirements or learning comparisons)
- **No New Languages in 2025**: No Zig, no OCaml, no Haskell, no Kotlin
- **Exception**: TypeScript for frontend work if absolutely necessary
- **Why**: Depth beats breadth. Mastery requires focus. Staff engineers are deep experts in their domain, not generalists.

### S4: Ship Ugly, Iterate Publicly
- **Anti-perfectionism protocol**:
  - Blog posts published as "living documents" (can update later)
  - Code shipped as v0.1.0 with known limitations documented
  - Talks submitted even if outline isn't perfect
  - PRs opened as WIP/draft to get early feedback
- **Rules**:
  - If 70% done, ship it
  - Perfect is the enemy of done
  - Feedback beats speculation
- **Why**: Perfectionism is my biggest enemy. Shipping imperfect work is better than perfect work never shipped.

### S5: Focus on One Project at a Time
- **Work in Progress (WIP) Limit**: 1 major project maximum
- **Definition of "Done"**:
  - Code merged/released
  - Blog post published
  - Talk given
  - Project archived or in maintenance mode
- **New Project Protocol**:
  - Can only start after current project reaches "Done"
  - Exception: Critical bug fixes or time-sensitive opportunities
- **Why**: Context switching kills momentum. Finishing one thing builds credibility more than starting ten things.

### S6: Consistent Publishing Schedule
- **Blog**: 1st of every month (draft by 25th of previous month)
- **OSS Contributions**: At least 4 PRs per month (1 per week)
- **Code Reviews**: Review 3 community PRs for every PR I open
- **Accountability**: Share monthly progress updates on blog
- **Why**: Consistency compounds. Regular output builds reputation and makes work visible.

---

## Stack

### Primary Technologies (Focus 90% of effort here)
- **Language**: Rust
- **Async Runtime**: Tokio
- **Web Frameworks**: Axum, Actix-Web
- **Database**: PostgreSQL (SQLx for Rust)
- **Testing**: Cargo test, Criterion (benchmarks), Proptest (property tests)
- **Observability**: Tracing, OpenTelemetry
- **CLI**: Clap v4
- **Serialization**: Serde

### Secondary Technologies (Only when required)
- **Language**: Go (work requirements only)
- **Frontend**: TypeScript, React (minimal, only if building full-stack tools)
- **Databases**: Redis, SQLite (for specific use cases)
- **Infrastructure**: Docker, Kubernetes (deployment, not development focus)
- **CI/CD**: GitHub Actions

### Explicitly Avoiding in 2025
- **No new languages**: Zig, Haskell, OCaml, Elixir, Kotlin
- **No new frameworks**: Unless critical for specific project goal
- **No "shiny object syndrome"**: Stay in Rust ecosystem

---

## Failure Patterns

### FP1: Tutorial Hell
- **Description**: Watching YouTube tutorials, reading books, taking courses instead of building real projects
- **Triggers**:
  - Feeling stuck on a problem
  - Seeing exciting new technology
  - Imposter syndrome ("I need to learn more before I start")
- **Manifestation**:
  - Saving articles to "read later" lists that never get read
  - Buying Udemy courses that sit at 10% completion
  - Following along with tutorials but not applying concepts
  - Building toy examples instead of real projects
- **Cost**: Months of "learning" without shipping anything real
- **Counter-Strategy**:
  - "Learn by doing" rule: Only read docs when blocked on actual project
  - Delete "read later" lists every month
  - If watching tutorial, must build own version immediately after
  - No courses allowed unless tied to specific project goal

### FP2: Shiny Object Syndrome (Context Switching)
- **Description**: Starting new projects before finishing current ones, jumping between technologies
- **Triggers**:
  - Seeing trending repo on Hacker News
  - Colleague mentions cool new framework
  - Getting bored with current project
  - Hitting difficult bug (running away from problems)
- **Manifestation**:
  - 20+ half-finished repos on GitHub
  - Trying Zig one week, OCaml the next, then going back to Rust
  - Starting blog posts but never publishing
  - Submitting conference talk proposals but not preparing talks
- **Cost**: No finished work to show, no depth in any technology, reputation as "dabbler" not "expert"
- **Counter-Strategy**:
  - WIP limit of 1 (enforced via this Telos file)
  - "Finish or kill" rule: Archive unfinished projects after 30 days of inactivity
  - New idea? Add to backlog, don't start immediately
  - Use `tm dump` to capture ideas without acting on them

### FP3: Perfectionism Paralysis
- **Description**: Refusing to ship until code/writing/talks are "perfect", over-engineering solutions
- **Triggers**:
  - Fear of judgment from peers
  - Comparing my work to established experts
  - Imagining negative feedback before getting any feedback
  - Thinking "I need to add just one more feature"
- **Manifestation**:
  - Blog posts sitting in drafts for months
  - Refactoring code endlessly instead of shipping
  - Not submitting conference talks because outline isn't perfect
  - Deleting code/writing and starting over
  - Adding features to side projects that no one asked for
- **Cost**: Nothing ships, no feedback loop, no growth, no credibility building
- **Counter-Strategy**:
  - "70% done is shipped" rule
  - Blog posts published as v0.1 (can update later)
  - Code shipped with limitations documented in README
  - Get feedback early (WIP PRs, draft posts to friends)
  - Time-box work (4 hours max per blog post first draft)

### FP4: Breadth Over Depth
- **Description**: Trying to learn everything instead of mastering one thing, wanting to be "full-stack" or "polyglot"
- **Triggers**:
  - Job postings asking for 10 different technologies
  - Feeling behind peers who know different tech stacks
  - FOMO from not knowing the "hot new framework"
  - Thinking "I should learn X to be more employable"
- **Manifestation**:
  - Switching between Rust, Go, Zig, Haskell monthly
  - Starting frontend projects when backend is my strength
  - Learning DevOps when I should focus on coding
  - Reading about databases, distributed systems, networking, compilers all at once
- **Cost**: Shallow knowledge of many things, deep expertise in nothing, not stand-out in any area
- **Counter-Strategy**:
  - "One Language Rule" (Rust in 2025)
  - Staff engineers are deep experts, not generalists
  - Say no to learning opportunities outside Rust ecosystem
  - Master async Rust, unsafe Rust, performance optimization, FFI

### FP5: No External Accountability
- **Description**: Working in isolation without deadlines, feedback, or public commitments
- **Triggers**:
  - Solo side projects
  - No boss for personal work
  - Fear of public failure
- **Manifestation**:
  - Projects that drag on for months
  - No one knows what I'm working on
  - Easy to abandon projects without consequences
  - No pressure to finish
- **Cost**: Low completion rate, no momentum, no external motivation
- **Counter-Strategy**:
  - Monthly blog updates on progress
  - Commit to deadlines publicly (blog, Twitter)
  - Join Rust community (Discord, forums) and share WIP
  - Find accountability buddy (another engineer with similar goals)
  - Use this Telos file and `tm` to track and score ideas

### FP6: Consumption Over Creation
- **Description**: Reading blogs, watching talks, browsing GitHub instead of creating
- **Triggers**:
  - Procrastinating on hard work
  - Feeling tired or low energy
  - "Research" phase that never ends
- **Manifestation**:
  - Hours on Hacker News daily
  - Reading every Rust blog post
  - Watching conference talks instead of preparing my own
  - Bookmarking repos instead of contributing
- **Cost**: Time wasted, no output, no growth
- **Counter-Strategy**:
  - "Create before consume" rule: Must write/code before reading
  - Limit HN/Reddit to 15 minutes per day (use Freedom app)
  - If watching talk, must take notes and publish summary
  - Track creation hours vs consumption hours weekly

---

## Missions

### M1: Become a Rust Ecosystem Authority
- **Focus**: Deep expertise in Rust async programming, systems-level performance, and production deployments
- **Actions**:
  - Contribute to Tokio, Axum, SQLx, or similar core libraries
  - Write advanced Rust content (not beginner tutorials)
  - Build production tools that solve real problems
  - Answer questions on Rust forums, /r/rust, Discord
  - Review PRs in Rust projects
- **Success Looks Like**:
  - Rust community members recognize my name
  - My blog posts are referenced in Rust discussions
  - Maintainers ask for my input on design decisions
  - Conference talks on Rust topics
- **Timeframe**: 2025 full year

### M2: Build a Technical Writing Portfolio
- **Focus**: Publishing technical content consistently to build credibility and audience
- **Actions**:
  - 12 blog posts in 2025 (1 per month)
  - Deep dives, not surface-level tutorials
  - Real-world problems and solutions
  - Performance analysis, debugging stories, architecture decisions
- **Success Looks Like**:
  - Blog with 500+ monthly readers
  - Posts shared on HN, Reddit, Twitter
  - Inbound messages from readers
  - Blog cited in job promotion packet
- **Timeframe**: 2025 full year

### M3: Demonstrate Technical Leadership
- **Focus**: Leading projects, mentoring others, influencing technical direction
- **Actions**:
  - Lead 2 cross-team initiatives at work
  - Mentor 2 junior engineers (1-on-1s, code reviews, career advice)
  - Write design docs for major features
  - Present tech talks internally (brown bags, lunch & learns)
- **Success Looks Like**:
  - Peers ask for my technical opinion
  - Mentees show measurable growth
  - Design docs approved and implemented
  - Recognized as "go-to person" for Rust/backend topics
- **Timeframe**: 2025 full year

### M4: Build Production-Quality Portfolio
- **Focus**: Ship side projects that demonstrate Staff-level engineering
- **Actions**:
  - Build 2 production-grade projects (not toys)
  - Document architecture, testing, deployment
  - Handle errors gracefully, log properly, monitor
  - Gather real user feedback and iterate
- **Success Looks Like**:
  - Projects used by actual users (not just me)
  - GitHub stars, issues, PRs from community
  - Referenced in job interviews/promotion discussions
  - Demonstrates end-to-end ownership
- **Timeframe**: 2025 full year

---

## Challenges

### C1: Balancing Full-Time Job and Side Projects
- **Problem**: 40-50 hour work weeks leave limited energy for OSS, blogging, conferences
- **Strategy**:
  - Early morning deep work (6am-10am) before work
  - Weekends reserved for side projects (at least 1 full day)
  - Use PTO strategically for conference prep or major pushes
  - Negotiate "OSS Fridays" with manager (1 day per month for contributions)
- **Guardrails**:
  - No side work during work hours
  - No burnout (track energy levels, take breaks)
  - Decline non-essential work meetings when possible

### C2: Imposter Syndrome When Publishing
- **Problem**: Fear that my content isn't good enough, that experts will criticize me
- **Strategy**:
  - "Learn in public" mindset: Share journey, not just achievements
  - Every expert was once a beginner
  - Negative feedback is rare; most people are supportive
  - Frame posts as "Here's what I learned" not "Here's what you should do"
- **Guardrails**:
  - Publish even when scared
  - Disable comments if anxiety is too high
  - Focus on helping one person, not impressing everyone

### C3: Maintaining Momentum Through Setbacks
- **Problem**: Rejected conference talks, ignored blog posts, slow OSS progress can kill motivation
- **Strategy**:
  - Celebrate small wins (PR merged, post published, not just outcomes)
  - Setbacks are data, not judgments
  - Keep long-term view (building for 2026 promotion, not just 2025 wins)
  - Find community support (accountability buddy, Rust Discord)
- **Guardrails**:
  - Monthly reflection on progress (even small)
  - Don't quit after one rejection
  - Revisit this Telos when motivation dips

### C4: Avoiding Tutorial Hell and Consumption Traps
- **Problem**: Easy to fall back into watching tutorials, reading blogs, "researching" instead of building
- **Strategy**:
  - "Build first, learn when stuck" rule
  - Time-box consumption (15 min HN per day max)
  - Track creation hours vs consumption hours
  - Use blocking apps (Freedom, Cold Turkey) during deep work
- **Guardrails**:
  - No new courses/books unless tied to active project
  - Reading docs is OK only when implementing something
  - Delete "read later" lists monthly

### C5: Staying Focused on Rust (Not Chasing Shiny Objects)
- **Problem**: Constant temptation to try Zig, Haskell, OCaml, etc. when Rust gets hard
- **Strategy**:
  - "One Language Rule" is non-negotiable in 2025
  - When tempted, add to backlog for 2026 review
  - Deep expertise requires sustained focus
  - Staff engineers are known for depth, not breadth
- **Guardrails**:
  - No new language repos on GitHub in 2025
  - Go is allowed only for work requirements
  - Use `tm dump` to capture ideas without acting

---

## Domain Keywords

**For filtering and scoring ideas**

- Rust
- Systems programming
- Distributed systems
- Async programming
- Performance optimization
- Tokio
- Axum
- Backend engineering
- Database internals
- Open source
- Technical writing
- Public speaking
- Conference talks
- Mentorship
- Technical leadership
- Production engineering
- Observability
- Testing strategies
- Architecture design
- Staff engineer
- Principal engineer
- Career growth

---

## Anti-Keywords

**Ideas containing these are likely distractions**

- Tutorial
- Course
- Beginner
- Learn X in Y days
- Full-stack (unless Rust backend focus)
- Frontend frameworks (unless minimal TS for tooling)
- DevOps certification
- AWS certification (unless critical for work)
- Mobile development
- Game development
- Web3/blockchain
- New programming language (except Rust)
- "Quick project"
- "Fun side project"
- "Portfolio piece" (unless production-quality)

---

## Success Metrics

### Quarterly Check-ins

**Q1 2025 (Jan-Mar)**
- [ ] 3 blog posts published
- [ ] 12+ OSS PRs merged
- [ ] 1 conference talk submitted
- [ ] Project 1 started and scoped
- [ ] 40+ hours deep work logged

**Q2 2025 (Apr-Jun)**
- [ ] 3 blog posts published
- [ ] 12+ OSS PRs merged
- [ ] Project 1 shipped (v0.1.0 minimum)
- [ ] 1 local meetup talk delivered
- [ ] 50+ hours deep work logged

**Q3 2025 (Jul-Sep)**
- [ ] 3 blog posts published
- [ ] 12+ OSS PRs merged
- [ ] Project 2 started and scoped
- [ ] 1 conference talk submitted (if Q1 not accepted)
- [ ] Blog at 300+ monthly visitors

**Q4 2025 (Oct-Dec)**
- [ ] 3 blog posts published
- [ ] 14+ OSS PRs merged
- [ ] Project 2 shipped
- [ ] 1 conference talk delivered
- [ ] Blog at 500+ monthly visitors
- [ ] Internal promotion packet drafted

### Year-End 2025 Success Criteria
- 12+ blog posts published
- 50+ OSS PRs merged
- 2 side projects shipped
- 1-2 conference talks delivered
- Blog averaging 500+ monthly visitors
- Recognized contributor in Rust community
- Internal promotion packet submitted (for 2026 H1 review)

---

## Notes

This Telos is a living document. I'll review it monthly and update as needed. The key is not perfection—it's focus, consistency, and shipping.

**My mantra for 2025**: Deep work. Ship ugly. Learn in public. Stay in Rust. Build for Staff.
