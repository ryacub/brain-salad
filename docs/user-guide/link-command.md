# Link Command: Managing Idea Relationships

## Table of Contents

- [Overview](#overview)
- [Why Use Relationships?](#why-use-relationships)
- [Relationship Types](#relationship-types)
- [Commands](#commands)
- [Common Workflows](#common-workflows)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [Examples](#examples)
- [Quick Start Cheat Sheet](#quick-start-cheat-sheet)

## Overview

The `link` command helps you create and manage relationships between your ideas. By linking ideas together, you can:

- **Track dependencies** - Know what needs to happen first
- **Organize ideas into hierarchies** - Break projects into tasks
- **Find related ideas quickly** - Discover connections
- **Understand how ideas connect** - Visualize your idea network
- **Identify blockers and duplicates** - Manage obstacles and avoid redundancy

The link command provides five subcommands:

| Command | Purpose |
|---------|---------|
| `create` | Create a new relationship between two ideas |
| `list` | List all relationships for an idea |
| `show` | Show related ideas with full details |
| `remove` | Remove a relationship |
| `path` | Find dependency paths between ideas |

## Why Use Relationships?

### Real-World Scenarios

**1. Project Planning**

Break down a big project into smaller, manageable tasks. Link subtasks to the main project to keep everything organized.

```bash
# Create project idea
tm dump "Build recipe sharing app"
# Returns ID: proj-abc123

# Create subtasks
tm dump "Design user interface"       # ui-def456
tm dump "Set up database"             # db-ghi789
tm dump "Build API endpoints"         # api-jkl012

# Link subtasks to project
tm link create ui-def456 proj-abc123 part_of
tm link create db-ghi789 proj-abc123 part_of
tm link create api-jkl012 proj-abc123 part_of
```

**2. Dependency Tracking**

Know which ideas must be completed before others can start.

```bash
# API depends on database being ready
tm link create api-jkl012 db-ghi789 depends_on

# UI depends on API being ready
tm link create ui-def456 api-jkl012 depends_on
```

**3. Knowledge Organization**

Group related ideas together for easier discovery and context.

```bash
# Link related authentication ideas
tm link create auth-login auth-session related_to
tm link create auth-oauth auth-login similar_to
```

**4. Duplicate Detection**

Mark duplicate ideas to avoid working on the same thing twice.

```bash
# Mark as duplicates
tm link create idea-001 idea-002 duplicate
```

**5. Blocker Management**

Track what's preventing progress on your ideas.

```bash
# Deployment blocked by pending security audit
tm link create deploy-prod security-audit blocked_by
```

## Relationship Types

Telos Matrix supports 9 different relationship types. Each type has a specific meaning and direction.

### Dependency Types

#### `depends_on`

**When to use:** Idea A cannot start until Idea B is complete

**Direction:** Source → Target (Source depends on Target)

**Example:**
- "Launch website" depends_on "Set up hosting"
- "Write Chapter 3" depends_on "Finish Chapter 2"
- "Deploy to production" depends_on "Pass all tests"

**How it works:**
```bash
tm link create launch-website setup-hosting depends_on
```
This means: "launch-website depends on setup-hosting" (can't launch until hosting is set up)

---

#### `blocked_by`

**When to use:** Idea A is blocked by an external factor or waiting on something

**Direction:** Source is blocked by Target

**Example:**
- "Deploy to production" blocked_by "Security audit pending"
- "Start user testing" blocked_by "Waiting for approvals"
- "Implement feature X" blocked_by "API not ready"

**How it works:**
```bash
tm link create deploy-prod security-audit blocked_by
```
This means: "deploy-prod is blocked by security-audit"

**Note:** `blocked_by` is similar to `depends_on` but emphasizes that something is actively preventing progress, rather than just a prerequisite.

---

#### `blocks`

**When to use:** Idea A prevents Idea B from proceeding

**Direction:** Source blocks Target

**Example:**
- "Critical bug fix" blocks "New feature development"
- "Infrastructure migration" blocks "App deployment"
- "Database schema change" blocks "API updates"

**How it works:**
```bash
tm link create bug-fix new-feature blocks
```
This means: "bug-fix blocks new-feature" (must fix bug before adding features)

**Note:** `blocks` is the inverse perspective of `blocked_by`. Use whichever makes more sense for your workflow.

---

### Hierarchy Types

#### `part_of`

**When to use:** Idea A is a component or subtask of Idea B

**Direction:** Source is part of Target

**Example:**
- "User authentication" part_of "User management system"
- "Login page" part_of "Website redesign"
- "Database setup" part_of "Backend infrastructure"

**How it works:**
```bash
tm link create login-page website-redesign part_of
```
This means: "login-page is part of website-redesign"

---

#### `parent`

**When to use:** Idea A is the parent of Idea B (A contains B)

**Direction:** Source has Target as a child

**Example:**
- "Mobile app" parent "Login screen"
- "Q4 2025 Goals" parent "Launch product"
- "Website" parent "Contact page"

**How it works:**
```bash
tm link create mobile-app login-screen parent
```
This means: "mobile-app is the parent of login-screen"

**Note:** `parent` is the inverse of `child`. Use `part_of` if you prefer thinking bottom-up (task → project), or `parent/child` if you prefer top-down (project → tasks).

---

#### `child`

**When to use:** Idea A is a child of Idea B (B contains A)

**Direction:** Source is a child of Target

**Example:**
- "Login screen" child "Mobile app"
- "Launch product" child "Q4 2025 Goals"

**How it works:**
```bash
tm link create login-screen mobile-app child
```
This means: "login-screen is a child of mobile-app"

---

### Association Types

#### `related_to`

**When to use:** Ideas are connected but don't have a strict dependency or hierarchy

**Direction:** Bidirectional (symmetric)

**Example:**
- "User authentication" related_to "Session management"
- "API design" related_to "Database schema"
- "Mobile app" related_to "Web app"

**How it works:**
```bash
tm link create user-auth session-mgmt related_to
```
This means: "user-auth and session-mgmt are related"

**Note:** This is a symmetric relationship, meaning it works both ways. There's no strict source/target distinction.

---

#### `similar_to`

**When to use:** Ideas are similar in approach, goal, or content, but still distinct

**Direction:** Bidirectional (symmetric)

**Example:**
- "Mobile app in React Native" similar_to "Web app in React"
- "GraphQL API" similar_to "REST API"
- "User dashboard v1" similar_to "User dashboard v2"

**How it works:**
```bash
tm link create mobile-app web-app similar_to
```
This means: "mobile-app is similar to web-app"

**Note:** Use this when ideas share characteristics but aren't duplicates. Good for comparing approaches or tracking iterations.

---

#### `duplicate`

**When to use:** Two ideas are essentially the same thing

**Direction:** Bidirectional (symmetric)

**Example:**
- "Add search functionality" duplicate "Implement search feature"
- "Build testing framework" duplicate "Create test suite"

**How it works:**
```bash
tm link create idea-001 idea-002 duplicate
```
This means: "idea-001 and idea-002 are the same idea"

**Note:** Mark ideas as duplicates to avoid working on the same thing twice. You can then archive one of them.

---

### Visual Relationship Guide

```
DEPENDENCIES:
depends_on:    A ──→ B   (A needs B to complete first)
blocked_by:    A ⊗─→ B   (B is blocking A)
blocks:        A ──⊗ B   (A prevents B)

HIERARCHIES:
part_of:       A ⊂ B    (A is inside/part of B)
parent:        A ⊃ B    (A contains B)
child:         A ⊂ B    (A is contained by B)

ASSOCIATIONS:
related_to:    A ←─→ B  (A and B are connected)
similar_to:    A ≈ B    (A and B are alike)
duplicate:     A = B    (A and B are the same)
```

## Commands

### Quick Reference

| Command | Purpose | Example |
|---------|---------|---------|
| `link create` | Create a relationship | `tm link create abc123 def456 depends_on` |
| `link list` | List all relationships for an idea | `tm link list abc123` |
| `link show` | Show related ideas with details | `tm link show abc123 --type depends_on` |
| `link remove` | Remove a relationship | `tm link remove rel789` |
| `link path` | Find paths between ideas | `tm link path abc123 xyz789` |

---

### `link create`

**Purpose:** Create a new relationship between two ideas

**Syntax:**
```bash
tm link create <source-id> <target-id> <type> [flags]
```

**Arguments:**
- `source-id`: The ID of the source idea
- `target-id`: The ID of the target idea
- `type`: Relationship type (see [Relationship Types](#relationship-types))

**Flags:**
- `--no-confirm`: Skip confirmation prompt

**How it works:**

1. You provide two idea IDs and a relationship type
2. The command shows you a preview of what will be linked
3. You confirm (unless using `--no-confirm`)
4. The relationship is created

**Examples:**

```bash
# Mark that "build API" depends on "design database"
tm link create api-abc123 db-def456 depends_on

# Mark ideas as duplicates (skip confirmation)
tm link create idea-001 idea-002 duplicate --no-confirm

# Create parent-child relationship for project breakdown
tm link create project-main subtask-001 parent

# Link related authentication ideas
tm link create oauth-impl session-mgmt related_to

# Mark blocker
tm link create deploy-prod security-audit blocked_by
```

**Sample Output:**

```
Creating relationship:
  Source: [api-abc1] Build REST API endpoints
  Target: [db-def45] Design database schema
  Type: depends_on

Continue? (y/n): y
✓ Relationship created successfully (ID: rel-ghi789)
```

**Common Mistakes:**

❌ **Mixing up source and target for `depends_on`**

If Task A depends on Task B, then:
- CORRECT: `tm link create taskA taskB depends_on`
- This means: "taskA depends on taskB"

❌ **Using the wrong relationship type**

- Use `depends_on` for "must complete before"
- Use `related_to` for "these are connected but no dependency"
- Use `part_of` for "this is a subtask of that"

❌ **Trying to create self-referencing relationship**

```bash
# This will fail:
tm link create abc123 abc123 depends_on
# Error: cannot create relationship from idea to itself
```

---

### `link list`

**Purpose:** List all relationships for an idea (both incoming and outgoing)

**Syntax:**
```bash
tm link list <idea-id>
```

**Arguments:**
- `idea-id`: The ID of the idea to show relationships for

**How it works:**

Displays two sections:
1. **Outgoing relationships** - Where this idea is the source
2. **Incoming relationships** - Where this idea is the target

**Examples:**

```bash
# List all relationships for an idea
tm link list abc123

# Get idea ID first, then list relationships
tm review pending | head -1  # Copy the ID
tm link list <copied-id>
```

**Sample Output:**

```
🔗 Relationships for idea: [api-abc1]
   Build REST API endpoints

Outgoing (where this idea is the source):
  1. depends_on → [db-def45] Design database schema
     ID: rel-001 | Created: 2025-11-19 10:30

  2. part_of → [proj-xyz] Recipe sharing app
     ID: rel-002 | Created: 2025-11-19 10:25

Incoming (where this idea is the target):
  1. depends_on ← [ui-ghi78] Build user interface
     ID: rel-003 | Created: 2025-11-19 10:35

  2. related_to ← [auth-jkl] Authentication service
     ID: rel-004 | Created: 2025-11-19 10:40

Total: 4 relationships
```

**Understanding the Output:**

- **Outgoing**: This idea depends on, is part of, or relates to other ideas
- **Incoming**: Other ideas depend on, are part of, or relate to this idea
- **Relationship ID**: Use this ID with `link remove` to delete the relationship
- **Direction arrows**:
  - `→` means this idea points to the related idea
  - `←` means the related idea points to this idea

---

### `link show`

**Purpose:** Show related ideas with full idea details (not just relationships)

**Syntax:**
```bash
tm link show <idea-id> [flags]
```

**Arguments:**
- `idea-id`: The ID of the idea to show related ideas for

**Flags:**
- `--type <relationship-type>`: Filter by specific relationship type

**How it works:**

Unlike `link list` which shows relationships, `link show` displays the full content and details of related ideas. This is useful when you want to see what the related ideas actually contain.

**Examples:**

```bash
# Show all related ideas
tm link show abc123

# Show only dependencies
tm link show abc123 --type depends_on

# Show only subtasks (things that are part of this idea)
tm link show abc123 --type part_of

# Show duplicates
tm link show abc123 --type duplicate
```

**Sample Output:**

```
🔗 Related ideas for: [api-abc1]
   Build REST API endpoints

1. Design database schema
   ID: db-def45 | Status: pending 📊 8.5/10
   Create PostgreSQL schema with tables for users, recipes, and ingredients. Include proper indexes and foreign key constraints.
   Created: 2025-11-19 09:15

2. Build user interface
   ID: ui-ghi78 | Status: in-progress 📊 7.2/10
   Create React-based UI for the recipe app with responsive design and dark mode support.
   Created: 2025-11-19 11:20

3. Authentication service
   ID: auth-jkl | Status: pending 📊 6.8/10
   Implement JWT-based authentication with OAuth2 support for Google and GitHub.
   Created: 2025-11-19 10:00

Found 3 related ideas
```

**When to use `link show` vs `link list`:**

- Use `link list` when you want to see the **structure** of relationships
- Use `link show` when you want to see the **content** of related ideas

---

### `link remove`

**Purpose:** Remove a relationship between ideas

**Syntax:**
```bash
tm link remove <relationship-id> [flags]
```

**Arguments:**
- `relationship-id`: The ID of the relationship to remove (get this from `link list`)

**Flags:**
- `--no-confirm`: Skip confirmation prompt

**How it works:**

1. You provide the relationship ID
2. The command shows you what will be removed
3. You confirm (unless using `--no-confirm`)
4. The relationship is deleted

**Important:** This only removes the **relationship**, not the ideas themselves. The ideas remain in your database.

**Examples:**

```bash
# First, list relationships to get the relationship ID
tm link list abc123
# Note the relationship ID (e.g., rel-ghi789)

# Remove the relationship
tm link remove rel-ghi789

# Remove without confirmation
tm link remove rel-ghi789 --no-confirm
```

**Sample Output:**

```
Removing relationship:
  ID: rel-ghi78
  [api-abc1] Build REST API endpoints
    depends_on →
  [db-def45] Design database schema

Are you sure? (y/n): y
✓ Relationship removed successfully
```

**Common Questions:**

**Q: Can I undo a removal?**

A: No, relationship removal is permanent. You'll need to recreate it with `link create` if removed by mistake.

**Q: Does removing a relationship delete the ideas?**

A: No, only the relationship link is deleted. Both ideas remain in your database.

**Q: How do I find the relationship ID?**

A: Use `tm link list <idea-id>` to see all relationships and their IDs.

---

### `link path`

**Purpose:** Find dependency paths between two ideas using breadth-first search

**Syntax:**
```bash
tm link path <source-id> <target-id> [flags]
```

**Arguments:**
- `source-id`: The starting idea ID
- `target-id`: The ending idea ID

**Flags:**
- `--max-depth <n>`: Maximum path length (default: 3)

**How it works:**

The command searches for all paths that connect the source idea to the target idea by following relationships. It uses a breadth-first search algorithm to find the shortest paths first.

This is useful for:
- Understanding dependency chains
- Finding how ideas are connected
- Discovering indirect relationships
- Analyzing complex project structures

**Examples:**

```bash
# Find paths between two ideas
tm link path start-idea end-idea

# Find paths with longer search depth
tm link path start-idea end-idea --max-depth 5

# Find how a subtask connects to the main project
tm link path subtask-id project-id --max-depth 2
```

**Sample Output:**

```
🔍 Finding paths from [ui-abc12] to [proj-xyz]...

Path 1 (2 hops):
  [ui-abc12] Build user interface
    → depends_on →
  [api-def45] Build REST API
    → part_of →
  [proj-xyz] Recipe sharing app

Path 2 (3 hops):
  [ui-abc12] Build user interface
    → related_to →
  [design-g] Design system
    → part_of →
  [brand-hi] Branding project
    → related_to →
  [proj-xyz] Recipe sharing app

Found 2 path(s)
```

**Understanding the Output:**

- **Hops**: Number of relationships in the path
- **Path**: Shows each idea and the relationship connecting to the next
- **Multiple paths**: If there are multiple ways to connect ideas, all are shown

**When No Path is Found:**

```
🔍 Finding paths from [abc123] to [xyz789]...

❌ No path found between [abc123] and [xyz789]

💡 Try linking ideas that might connect these two concepts
```

**Performance Tip:**

If you have a large idea network, use a smaller `--max-depth` to speed up the search:

```bash
# Faster search with depth limit
tm link path abc123 xyz789 --max-depth 2
```

## Common Workflows

### Workflow 1: Breaking Down a Project

**Scenario:** You have a big project and want to break it into smaller tasks.

**Steps:**

1. **Create the main project idea:**

```bash
tm dump "Build a recipe sharing app"
# Note the ID, e.g., proj-abc123
```

2. **Create sub-tasks:**

```bash
tm dump "Design user interface"      # ui-def456
tm dump "Set up database"            # db-ghi789
tm dump "Build API endpoints"        # api-jkl012
tm dump "Write documentation"        # docs-mno345
```

3. **Link tasks to project:**

```bash
tm link create ui-def456 proj-abc123 part_of
tm link create db-ghi789 proj-abc123 part_of
tm link create api-jkl012 proj-abc123 part_of
tm link create docs-mno345 proj-abc123 part_of
```

4. **View the project structure:**

```bash
# List all relationships for the project
tm link list proj-abc123

# Or show related ideas with full details
tm link show proj-abc123 --type part_of
```

**Alternative Approach (Top-Down):**

If you prefer thinking from project → tasks, use `parent` instead:

```bash
tm link create proj-abc123 ui-def456 parent
tm link create proj-abc123 db-ghi789 parent
tm link create proj-abc123 api-jkl012 parent
```

---

### Workflow 2: Tracking Dependencies

**Scenario:** You need to know what order to tackle ideas.

**Steps:**

1. **Identify dependencies:**

Think about what must happen first:
- Database design must happen before API development
- API must be ready before UI can consume it
- Everything must be done before documentation

2. **Create dependency links:**

```bash
# API depends on database
tm link create api-jkl012 db-ghi789 depends_on

# UI depends on API
tm link create ui-def456 api-jkl012 depends_on

# Documentation depends on UI
tm link create docs-mno345 ui-def456 depends_on
```

3. **Find the full dependency chain:**

```bash
# See how documentation connects to database
tm link path docs-mno345 db-ghi789
```

Output:
```
Path 1 (3 hops):
  [docs-mno] Write documentation
    → depends_on →
  [ui-def45] Design user interface
    → depends_on →
  [api-jkl0] Build API endpoints
    → depends_on →
  [db-ghi78] Set up database
```

4. **View all dependencies for a task:**

```bash
# What does the UI depend on?
tm link list ui-def456
```

---

### Workflow 3: Finding Related Ideas

**Scenario:** You want to group similar ideas together for easier discovery.

**Steps:**

1. **Capture related ideas:**

```bash
tm dump "User authentication with JWT"       # auth-001
tm dump "Session management system"          # session-002
tm dump "OAuth2 integration"                 # oauth-003
tm dump "Password reset functionality"       # password-004
```

2. **Create relationship links:**

```bash
# Link authentication concepts
tm link create auth-001 session-002 related_to
tm link create auth-001 oauth-003 related_to
tm link create auth-001 password-004 related_to

# OAuth is similar to JWT auth
tm link create oauth-003 auth-001 similar_to
```

3. **Explore the network:**

```bash
# See all ideas related to authentication
tm link show auth-001

# Find specific relationship types
tm link show auth-001 --type related_to
```

---

### Workflow 4: Managing Blockers

**Scenario:** Some ideas are blocked by external factors.

**Steps:**

1. **Identify blockers:**

```bash
tm dump "Deploy to production"               # deploy-001
tm dump "Waiting for security audit"         # audit-002
tm dump "Need database credentials"          # creds-003
```

2. **Mark blocked relationships:**

```bash
# Deployment blocked by security audit
tm link create deploy-001 audit-002 blocked_by

# Deployment also blocked by missing credentials
tm link create deploy-001 creds-003 blocked_by
```

3. **Review blocked ideas:**

```bash
# See what's blocking deployment
tm link list deploy-001
```

4. **Remove blocker when resolved:**

```bash
# Once audit is complete, remove the blocker
tm link list deploy-001  # Get relationship ID
tm link remove rel-xyz789
```

---

### Workflow 5: Avoiding Duplicate Work

**Scenario:** You want to mark duplicate ideas to avoid redundant work.

**Steps:**

1. **Find potential duplicates:**

```bash
# Review recent ideas
tm review pending --limit 20
```

2. **Mark as duplicate:**

```bash
# These two are the same idea
tm link create idea-001 idea-002 duplicate
```

3. **Archive one of the duplicates:**

```bash
# Keep the better-scored one, archive the other
tm update idea-002 --status archived
```

4. **Review duplicates:**

```bash
# See if an idea has duplicates
tm link show idea-001 --type duplicate
```

---

### Workflow 6: Sprint Planning

**Scenario:** Planning a 2-week sprint with dependencies.

**Steps:**

1. **Create sprint container:**

```bash
tm dump "Sprint 12 - Nov 19 to Dec 2"        # sprint-012
```

2. **Create sprint tasks:**

```bash
tm dump "Implement user profiles"            # task-001
tm dump "Add profile photo upload"           # task-002
tm dump "Create settings page"               # task-003
tm dump "Write API documentation"            # task-004
```

3. **Link tasks to sprint:**

```bash
tm link create task-001 sprint-012 part_of
tm link create task-002 sprint-012 part_of
tm link create task-003 sprint-012 part_of
tm link create task-004 sprint-012 part_of
```

4. **Add dependencies:**

```bash
# Photo upload depends on profiles
tm link create task-002 task-001 depends_on

# Settings depends on profiles
tm link create task-003 task-001 depends_on

# Docs depend on all features being done
tm link create task-004 task-001 depends_on
tm link create task-004 task-002 depends_on
tm link create task-004 task-003 depends_on
```

5. **View sprint plan:**

```bash
# See all sprint tasks
tm link show sprint-012 --type part_of

# Check dependencies for each task
tm link list task-002
tm link list task-003
tm link list task-004
```

## Best Practices

### Choosing the Right Relationship Type

**Decision Tree:**

1. **Are they the same idea?**
   - Yes → Use `duplicate`

2. **Is one idea part of a larger idea?**
   - Yes, thinking bottom-up (task → project) → Use `part_of`
   - Yes, thinking top-down (project → task) → Use `parent` or `child`

3. **Does one idea need another to complete first?**
   - Yes, and it's a hard requirement → Use `depends_on`
   - Yes, and it's being actively blocked → Use `blocked_by`
   - Yes, and this idea is preventing another → Use `blocks`

4. **Are they similar but different?**
   - Yes, different approaches to same goal → Use `similar_to`

5. **Are they just related in some way?**
   - Yes → Use `related_to`

### When to Use Each Type

| Situation | Use This | Example |
|-----------|----------|---------|
| Breaking down projects | `part_of` or `parent`/`child` | Website → Login page, Home page |
| Task ordering | `depends_on` | "Deploy" depends on "Testing" |
| Same idea twice | `duplicate` | "Add search" and "Implement search" |
| Waiting on something | `blocked_by` | "Launch" blocked by "Legal approval" |
| Preventing progress | `blocks` | "Bug fix" blocks "New feature" |
| Loosely related | `related_to` | "User auth" related to "Session mgmt" |
| Similar approaches | `similar_to` | "Mobile app" similar to "Web app" |

### Tips for Effective Use

**1. Be Consistent**

Choose one pattern and stick with it:
- Either use `parent`/`child` OR `part_of` for hierarchies, not both
- Either use `depends_on` OR `blocked_by` for dependencies, based on your preference

**2. Keep It Simple**

Don't over-link:
- Link only meaningful relationships
- Too many links create noise and confusion
- Focus on the most important connections

**3. Use Path-Finding**

Let the tool find connections for you:
- Don't manually track long chains
- Use `tm link path` to discover how ideas connect
- This helps identify indirect relationships

**4. Review Regularly**

Clean up old relationships:
- Remove links when ideas are completed or archived
- Update relationships as plans change
- Review dependencies periodically

**5. Document Why**

Add context to your ideas:
- When creating an idea, include why it depends on another
- Use clear, descriptive idea content
- This helps future you understand your thinking

**6. Start Small**

Begin with simple relationships:
- Start with just `depends_on` and `part_of`
- Add other types as you get comfortable
- Don't feel obligated to use all 9 types

**7. Use Symmetric Relationships Wisely**

For `related_to`, `similar_to`, and `duplicate`:
- You only need to create the link once
- It automatically works in both directions
- No need to create reverse relationships

**8. Combine with Status Updates**

Track progress by updating idea status:
```bash
# When a dependency is complete
tm update db-ghi789 --status completed

# Then review what's now unblocked
tm link list api-jkl012
```

## Troubleshooting

### Common Issues

#### "Idea not found"

**Problem:** The idea ID doesn't exist in the database

**Solution:**

```bash
# List recent ideas to find the right ID
tm analyze --limit 20

# Search for specific content (if available)
tm review pending | grep "keyword"

# Or get idea details
tm review <partial-id>
```

**Tip:** You can use truncated IDs (first 8 characters) in most cases.

---

#### "Relationship already exists"

**Problem:** You've already created this exact relationship between these ideas

**Solution:**

- Check existing relationships:
  ```bash
  tm link list <idea-id>
  ```

- If it's truly a duplicate, you're done!
- If you want a different type, remove the old one first:
  ```bash
  tm link remove <rel-id>
  tm link create <source-id> <target-id> <new-type>
  ```

---

#### "Cannot create relationship from idea to itself"

**Problem:** Source and target IDs are the same

**Solution:**

- Verify you have two different idea IDs
- Double-check you didn't copy-paste the same ID twice
- Make sure you're not using a truncated ID that matches multiple ideas

---

#### "Invalid relationship type"

**Problem:** The relationship type you specified doesn't exist

**Solution:**

Valid types are:
- `depends_on`
- `related_to`
- `part_of`
- `parent`
- `child`
- `duplicate`
- `blocks`
- `blocked_by`
- `similar_to`

Check for typos:
- Use underscores, not hyphens: `depends_on` not `depends-on`
- Use lowercase: `depends_on` not `DEPENDS_ON`
- Use full names: `depends_on` not `depends`

---

#### Can't find relationship ID to remove

**Problem:** You want to remove a relationship but don't know its ID

**Solution:**

```bash
# List all relationships for the idea
tm link list <idea-id>

# Copy the relationship ID from the output
# It will be shown as "ID: rel-xyz789"

# Then remove it
tm link remove rel-xyz789
```

---

#### Path-finding is slow

**Problem:** `tm link path` is taking too long

**Solution:**

Reduce the search depth:
```bash
# Instead of default depth 3
tm link path abc123 xyz789 --max-depth 2
```

This is usually caused by:
- A very large number of relationships
- Many interconnected ideas
- Deep dependency chains

---

#### Too many relationships to manage

**Problem:** An idea has dozens of relationships and it's overwhelming

**Solution:**

1. **Filter by type:**
   ```bash
   # Only show dependencies
   tm link show <idea-id> --type depends_on
   ```

2. **Use hierarchies:**
   - Group related ideas under a parent
   - Link subtasks to the parent instead of to each other

3. **Clean up old relationships:**
   - Remove relationships to completed/archived ideas
   - Use `link remove` to delete outdated links

4. **Consider if you're over-linking:**
   - Not every connection needs a relationship
   - Focus on the most important dependencies

---

#### Created relationship in wrong direction

**Problem:** You created `A depends_on B` but meant `B depends_on A`

**Solution:**

```bash
# First, remove the incorrect relationship
tm link list A  # Get the relationship ID
tm link remove rel-xyz789

# Then create the correct relationship
tm link create B A depends_on
```

**Tip:** Always think carefully about direction for asymmetric relationships like `depends_on`, `blocks`, and `part_of`.

---

## Examples

### Example 1: Building a SaaS Product

**Scenario:** You're building a SaaS product and want to organize your entire roadmap.

**Step 1: Create the main project**

```bash
tm dump "TaskFlow - Project Management SaaS"
# ID: saas-main
```

**Step 2: Create major components**

```bash
tm dump "User authentication system"         # auth-sys
tm dump "Project dashboard"                  # dash-proj
tm dump "Task management module"             # task-mod
tm dump "Team collaboration features"        # team-feat
tm dump "Billing and subscriptions"          # billing
tm dump "Admin panel"                        # admin
```

**Step 3: Link components to main project**

```bash
tm link create auth-sys saas-main part_of
tm link create dash-proj saas-main part_of
tm link create task-mod saas-main part_of
tm link create team-feat saas-main part_of
tm link create billing saas-main part_of
tm link create admin saas-main part_of
```

**Step 4: Add dependencies**

```bash
# Dashboard depends on auth
tm link create dash-proj auth-sys depends_on

# Task module depends on auth and dashboard
tm link create task-mod auth-sys depends_on
tm link create task-mod dash-proj depends_on

# Team features depend on task module
tm link create team-feat task-mod depends_on

# Billing depends on auth
tm link create billing auth-sys depends_on

# Admin depends on everything else
tm link create admin auth-sys depends_on
tm link create admin dash-proj depends_on
tm link create admin task-mod depends_on
```

**Step 5: Create subtasks for auth system**

```bash
tm dump "Implement JWT authentication"       # jwt-impl
tm dump "Add OAuth2 providers"              # oauth-add
tm dump "Password reset flow"               # pwd-reset
tm dump "Email verification"                # email-ver

# Link to auth system
tm link create jwt-impl auth-sys part_of
tm link create oauth-add auth-sys part_of
tm link create pwd-reset auth-sys part_of
tm link create email-ver auth-sys part_of

# OAuth depends on JWT being done
tm link create oauth-add jwt-impl depends_on
```

**Step 6: View the structure**

```bash
# See all components of the SaaS
tm link show saas-main --type part_of

# See auth system subtasks
tm link show auth-sys --type part_of

# See what dashboard depends on
tm link list dash-proj

# Find path from admin to JWT implementation
tm link path admin jwt-impl
```

---

### Example 2: Research Project Organization

**Scenario:** You're conducting a research project and want to track different aspects.

**Step 1: Create research project**

```bash
tm dump "Research: Impact of AI on software development"
# ID: research-ai
```

**Step 2: Create research phases**

```bash
tm dump "Literature review"                  # lit-review
tm dump "Design research methodology"        # methodology
tm dump "Conduct interviews"                 # interviews
tm dump "Analyze survey data"                # survey-data
tm dump "Write research paper"               # paper
```

**Step 3: Link phases to project**

```bash
tm link create lit-review research-ai part_of
tm link create methodology research-ai part_of
tm link create interviews research-ai part_of
tm link create survey-data research-ai part_of
tm link create paper research-ai part_of
```

**Step 4: Add dependencies (research workflow)**

```bash
# Methodology depends on literature review
tm link create methodology lit-review depends_on

# Interviews depend on methodology
tm link create interviews methodology depends_on

# Survey analysis depends on methodology
tm link create survey-data methodology depends_on

# Paper depends on everything
tm link create paper lit-review depends_on
tm link create paper interviews depends_on
tm link create paper survey-data depends_on
```

**Step 5: Add related research ideas**

```bash
tm dump "Explore GPT impact on code quality"     # gpt-quality
tm dump "Study developer productivity metrics"   # dev-metrics
tm dump "Analyze adoption barriers"              # barriers

# Link to main research
tm link create gpt-quality research-ai related_to
tm link create dev-metrics research-ai related_to
tm link create barriers research-ai related_to

# Link related ideas to each other
tm link create gpt-quality dev-metrics similar_to
tm link create dev-metrics barriers related_to
```

**Step 6: Track blockers**

```bash
tm dump "Wait for IRB approval"              # irb-approval
tm link create interviews irb-approval blocked_by
```

**Step 7: Review progress**

```bash
# See all research components
tm link show research-ai

# Check what's needed before writing paper
tm link list paper

# Find path from paper to literature review
tm link path paper lit-review
```

---

### Example 3: Sprint Planning

**Scenario:** Planning a 2-week sprint with a team.

**Complete workflow:**

```bash
# Create sprint
tm dump "Sprint 15 - Authentication & Profile Features"
# ID: sprint-15

# Create user stories
tm dump "As a user, I want to log in with email/password"
# ID: story-login

tm dump "As a user, I want to reset my forgotten password"
# ID: story-reset

tm dump "As a user, I want to update my profile information"
# ID: story-profile

tm dump "As a user, I want to upload a profile photo"
# ID: story-photo

# Link stories to sprint
tm link create story-login sprint-15 part_of
tm link create story-reset sprint-15 part_of
tm link create story-profile sprint-15 part_of
tm link create story-photo sprint-15 part_of

# Create technical tasks
tm dump "Implement JWT authentication backend"
# ID: task-jwt

tm dump "Create login form UI"
# ID: task-login-ui

tm dump "Build password reset email service"
# ID: task-email

tm dump "Design profile settings page"
# ID: task-profile-ui

tm dump "Implement image upload API"
# ID: task-upload

# Link tasks to user stories
tm link create task-jwt story-login part_of
tm link create task-login-ui story-login part_of
tm link create task-email story-reset part_of
tm link create task-profile-ui story-profile part_of
tm link create task-upload story-photo part_of

# Add dependencies between tasks
tm link create task-login-ui task-jwt depends_on
tm link create task-email task-jwt depends_on
tm link create task-profile-ui task-jwt depends_on
tm link create task-upload task-profile-ui depends_on

# Add blocker
tm dump "Need S3 bucket for profile photos"
# ID: blocker-s3

tm link create task-upload blocker-s3 blocked_by

# View sprint structure
tm link show sprint-15 --type part_of

# Check task dependencies
tm link list task-upload

# Find critical path
tm link path task-upload task-jwt
```

This gives you a complete view of:
- Sprint goals (user stories)
- Implementation tasks
- Dependencies (what must be done first)
- Blockers (what's preventing progress)

---

## Quick Start Cheat Sheet

```bash
# CREATE A RELATIONSHIP
tm link create <source-id> <target-id> <type>

# Common relationship types:
tm link create taskA taskB depends_on     # TaskA depends on TaskB
tm link create subtask project part_of    # Subtask is part of project
tm link create ideaX ideaY related_to     # Ideas are related
tm link create idea1 idea2 duplicate      # Mark as duplicates
tm link create taskA blocker blocked_by   # TaskA blocked by blocker

# LIST ALL RELATIONSHIPS
tm link list <idea-id>

# SHOW RELATED IDEAS WITH DETAILS
tm link show <idea-id>
tm link show <idea-id> --type depends_on  # Filter by type

# REMOVE A RELATIONSHIP
tm link list <idea-id>                    # Get relationship ID
tm link remove <rel-id>

# FIND PATH BETWEEN IDEAS
tm link path <start-id> <end-id>
tm link path <start-id> <end-id> --max-depth 5

# SKIP CONFIRMATION PROMPTS
tm link create <source> <target> <type> --no-confirm
tm link remove <rel-id> --no-confirm
```

### Relationship Types Quick Reference

```
depends_on    - Source needs target to complete first
blocked_by    - Source is blocked by target
blocks        - Source blocks target
part_of       - Source is part of target (bottom-up)
parent        - Source is parent of target (top-down)
child         - Source is child of target (bottom-up)
related_to    - Ideas are connected (symmetric)
similar_to    - Ideas are similar (symmetric)
duplicate     - Ideas are the same (symmetric)
```

---

## Next Steps

- Learn about [Bulk Operations](../CLI_REFERENCE.md#dump) for managing many ideas
- Explore [Analytics](../CLI_REFERENCE.md#analytics) to analyze your idea network
- Check out the [Quick Reference Cheat Sheet](../quick-reference/link-cheatsheet.md)
- Try the [Getting Started Tutorial](../tutorials/getting-started-with-links.md)
- Read the [FAQ](../faq/link-command-faq.md) for common questions

---

**Last Updated:** 2025-11-19
**Version:** 1.0
**Command:** `tm link --help`
