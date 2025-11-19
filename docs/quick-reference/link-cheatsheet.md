# Link Command Cheat Sheet

Quick reference for the Telos Matrix link command.

## Commands

```bash
# CREATE RELATIONSHIP
tm link create <source-id> <target-id> <type> [--no-confirm]

# LIST RELATIONSHIPS
tm link list <idea-id>

# SHOW RELATED IDEAS
tm link show <idea-id> [--type <type>]

# REMOVE RELATIONSHIP
tm link remove <relationship-id> [--no-confirm]

# FIND PATH
tm link path <from-id> <to-id> [--max-depth N]
```

## Relationship Types

| Type | Direction | Meaning | Example |
|------|-----------|---------|---------|
| `depends_on` | A → B | A needs B first | Task A depends on Task B |
| `blocked_by` | A ⊗ B | B blocks A | Deploy blocked by audit |
| `blocks` | A ⊗ B | A prevents B | Bug blocks feature |
| `part_of` | A ⊂ B | A inside B | Task part of project |
| `parent` | A ⊃ B | A contains B | Project contains task |
| `child` | A ⊂ B | B contains A | Task child of project |
| `related_to` | A ↔ B | Connected | Auth related to sessions |
| `similar_to` | A ≈ B | Similar | Mobile similar to web |
| `duplicate` | A = B | Same idea | Duplicate entries |

### Symbols

```
→  Direction (one-way)
↔  Bidirectional (symmetric)
⊗  Blocking
⊂  Part of / Contained by
⊃  Contains
≈  Similar
=  Identical
```

## Common Patterns

### Project Breakdown

```bash
# Create project and tasks
tm dump "Build SaaS Product"          # proj-123
tm dump "User authentication"         # auth-456
tm dump "Payment integration"         # pay-789

# Link tasks to project (bottom-up)
tm link create auth-456 proj-123 part_of
tm link create pay-789 proj-123 part_of

# Alternative: top-down
tm link create proj-123 auth-456 parent
tm link create proj-123 pay-789 parent
```

### Task Dependencies

```bash
# B must complete before A
tm link create taskA taskB depends_on

# Chain dependencies
tm link create taskC taskB depends_on
tm link create taskB taskA depends_on

# View chain
tm link path taskC taskA
```

### Mark Duplicates

```bash
# Find duplicates
tm review pending | grep "search"

# Mark as duplicate
tm link create idea-001 idea-002 duplicate

# Archive one
tm update idea-002 --status archived
```

### Track Blockers

```bash
# Mark blocker
tm link create deployment audit blocked_by

# Remove when resolved
tm link list deployment  # Get rel-id
tm link remove rel-xyz
```

## Workflow Examples

### Sprint Planning

```bash
# 1. Create sprint
tm dump "Sprint 10"                   # sprint-10

# 2. Create and link stories
tm dump "User login story"            # story-1
tm link create story-1 sprint-10 part_of

# 3. Create and link tasks
tm dump "Build login API"             # task-1
tm link create task-1 story-1 part_of

# 4. Add dependencies
tm dump "Set up database"             # task-2
tm link create task-1 task-2 depends_on

# 5. Review
tm link show sprint-10
```

### Research Organization

```bash
# Create main research
tm dump "AI Impact Research"          # research-1

# Create phases
tm dump "Literature review"           # phase-1
tm dump "Data collection"             # phase-2
tm dump "Write paper"                 # phase-3

# Link phases
tm link create phase-1 research-1 part_of
tm link create phase-2 research-1 part_of
tm link create phase-3 research-1 part_of

# Add dependencies
tm link create phase-2 phase-1 depends_on
tm link create phase-3 phase-2 depends_on
```

## Tips

### Choosing Types

1. **Same idea?** → `duplicate`
2. **Part of larger idea?** → `part_of` or `parent`/`child`
3. **Must complete first?** → `depends_on`
4. **Being blocked?** → `blocked_by` or `blocks`
5. **Similar approaches?** → `similar_to`
6. **Just related?** → `related_to`

### Best Practices

- ✅ Use `depends_on` for hard dependencies
- ✅ Use `part_of` for hierarchies (or `parent`/`child`)
- ✅ Use `related_to` when relationship is unclear
- ✅ Mark duplicates to avoid redundant work
- ❌ Don't over-link everything
- ❌ Don't mix `parent`/`child` with `part_of`
- ❌ Don't create circular dependencies

### Symmetric Types

These work bidirectionally (create once):
- `related_to`
- `similar_to`
- `duplicate`

No need to create reverse relationships!

## Troubleshooting

| Error | Cause | Solution |
|-------|-------|----------|
| "Idea not found" | Invalid idea ID | Check ID with `tm review` |
| "Relationship already exists" | Duplicate relationship | Check with `tm link list` |
| "Cannot create relationship from idea to itself" | Same source and target | Use different IDs |
| "Invalid relationship type" | Typo in type | Use underscores: `depends_on` |
| Can't find relationship ID | Need to remove but no ID | Use `tm link list` to get ID |

## Quick Syntax Reference

### Create Relationships

```bash
tm link create api-123 db-456 depends_on
tm link create subtask project part_of
tm link create idea1 idea2 duplicate --no-confirm
```

### List and Show

```bash
tm link list abc123
tm link show abc123
tm link show abc123 --type depends_on
```

### Remove

```bash
tm link list abc123          # Get relationship ID
tm link remove rel-xyz789
tm link remove rel-xyz789 --no-confirm
```

### Find Paths

```bash
tm link path start-id end-id
tm link path start-id end-id --max-depth 5
```

## Output Examples

### Create Output

```
Creating relationship:
  Source: [api-abc1] Build REST API
  Target: [db-def45] Design database
  Type: depends_on

Continue? (y/n): y
✓ Relationship created successfully (ID: rel-ghi789)
```

### List Output

```
🔗 Relationships for idea: [api-abc1]
   Build REST API endpoints

Outgoing (where this idea is the source):
  1. depends_on → [db-def45] Design database schema
     ID: rel-001 | Created: 2025-11-19 10:30

Incoming (where this idea is the target):
  1. depends_on ← [ui-ghi78] Build user interface
     ID: rel-002 | Created: 2025-11-19 10:35

Total: 2 relationships
```

### Path Output

```
🔍 Finding paths from [ui-abc12] to [db-ghi78]...

Path 1 (2 hops):
  [ui-abc12] Build user interface
    → depends_on →
  [api-def45] Build REST API
    → depends_on →
  [db-ghi78] Design database

Found 1 path(s)
```

## See Also

- [Link Command User Guide](../user-guide/link-command.md) - Comprehensive documentation
- [Getting Started Tutorial](../tutorials/getting-started-with-links.md) - Step-by-step guide
- [FAQ](../faq/link-command-faq.md) - Common questions
- [CLI Reference](../CLI_REFERENCE.md) - All commands

---

**Version:** 1.0 | **Last Updated:** 2025-11-19
