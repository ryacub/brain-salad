# Link Command FAQ

Frequently asked questions about the Telos Matrix link command.

## Table of Contents

- [General Questions](#general-questions)
- [Relationship Types](#relationship-types)
- [Technical Questions](#technical-questions)
- [Troubleshooting](#troubleshooting)
- [Best Practices](#best-practices)

## General Questions

### Q: When should I use relationships?

**A:** Use relationships when ideas are connected in meaningful ways:

- When one idea depends on another
- When ideas are part of a larger project
- When you want to group related concepts
- When you need to track blockers
- When you've captured duplicate ideas

Don't over-link! Not every connection needs a relationship. Focus on the most important dependencies and hierarchies.

---

### Q: What's the difference between `depends_on` and `blocked_by`?

**A:** Both indicate dependencies, but from different perspectives:

**`depends_on`:**
- Indicates a prerequisite
- "I need this to complete first"
- Future-oriented: "When X is done, I can start"
- Example: "Deploy to production" depends_on "Pass all tests"

**`blocked_by`:**
- Indicates an active blocker
- "This is preventing me from proceeding"
- Present-oriented: "I'm stuck because of X"
- Example: "Deploy to production" blocked_by "Waiting for security audit"

Use whichever makes more sense for your workflow. They're similar enough that you can choose based on personal preference.

---

### Q: Should I use `part_of` or `parent`/`child`?

**A:** Choose one approach and stick with it:

**`part_of` (bottom-up thinking):**
- "This task is part of that project"
- Natural for task-focused workflows
- Example: `tm link create task project part_of`

**`parent`/`child` (top-down thinking):**
- "This project has these children"
- Natural for project-focused workflows
- Example: `tm link create project task parent`

Both achieve the same result, just from different perspectives. Pick whichever feels more natural to you.

**Don't mix them!** Mixing creates confusion in your relationship graph.

---

### Q: Can I link more than two ideas together?

**A:** Each relationship connects exactly two ideas, but one idea can have many relationships.

For example, a project can have multiple tasks:

```bash
# Project has 5 tasks
tm link create task1 project part_of
tm link create task2 project part_of
tm link create task3 project part_of
tm link create task4 project part_of
tm link create task5 project part_of
```

Then viewing the project shows all 5 tasks:

```bash
tm link show project --type part_of
```

---

### Q: What happens if I delete an idea that has relationships?

**A:** The relationships are also deleted automatically. This is by design to keep your database clean.

If you want to preserve the relationship structure, consider archiving the idea instead of deleting it:

```bash
tm update <idea-id> --status archived
```

---

### Q: Can I create bidirectional relationships?

**A:** Some relationship types are automatically bidirectional (symmetric):

- `related_to` - Works both ways
- `similar_to` - Works both ways
- `duplicate` - Works both ways

For these types, you only create the relationship once:

```bash
tm link create ideaA ideaB related_to
```

Now both `tm link list ideaA` and `tm link list ideaB` will show the relationship.

For non-symmetric types like `depends_on`, you need to create two separate relationships if you want both directions (though this would create a circular dependency, which you should avoid).

---

## Relationship Types

### Q: When should I use `related_to` vs `similar_to`?

**A:**

**`related_to`:**
- Ideas are connected but serve different purposes
- They work together or complement each other
- Example: "User authentication" related_to "Session management"
- Think: "These work together"

**`similar_to`:**
- Ideas are alike in approach, goal, or content
- Often different implementations of the same concept
- Example: "Mobile app" similar_to "Web app"
- Think: "These are different versions of the same thing"

---

### Q: What's the difference between `blocks` and `blocked_by`?

**A:** They're inverse perspectives of the same relationship:

**`blocks`:**
- Source prevents target
- "I'm blocking that"
- Example: `tm link create bug-fix new-feature blocks`
- Meaning: "This bug fix is blocking new feature development"

**`blocked_by`:**
- Source is blocked by target
- "That's blocking me"
- Example: `tm link create new-feature bug-fix blocked_by`
- Meaning: "New feature is blocked by the bug fix"

Choose based on which idea you're thinking from. They mean the same thing.

---

### Q: Should I use `duplicate` or just delete one idea?

**A:** Use `duplicate` first, then archive or delete:

```bash
# 1. Mark as duplicate
tm link create idea1 idea2 duplicate

# 2. Review both and decide which to keep
tm link show idea1 --type duplicate

# 3. Archive the one you don't want
tm update idea2 --status archived
```

This is better than immediate deletion because:
- You can review both ideas before deciding
- You maintain a record that they're the same
- You can recover the archived idea if needed

---

### Q: Can I create custom relationship types?

**A:** No, the relationship types are fixed. The 9 built-in types cover most use cases:

- `depends_on` - Dependencies
- `blocked_by` - Blockers
- `blocks` - Blocking
- `part_of` - Hierarchies (bottom-up)
- `parent` - Hierarchies (top-down)
- `child` - Hierarchies (bottom-up)
- `related_to` - General associations
- `similar_to` - Similar ideas
- `duplicate` - Same idea

If none fit perfectly, use `related_to` as a catch-all.

---

## Technical Questions

### Q: How does path-finding work?

**A:** The path-finding algorithm uses breadth-first search (BFS):

1. Starts at the source idea
2. Explores all directly connected ideas
3. Then explores ideas connected to those ideas
4. Continues until it finds the target or reaches max depth
5. Returns all found paths, shortest first

**Performance:**
- Fast for small to medium graphs (< 1000 ideas)
- Slower for large, highly-connected graphs
- Use `--max-depth` to limit search depth

**Example:**

```bash
# Find paths up to 3 relationships deep (default)
tm link path start end

# Limit depth for faster search
tm link path start end --max-depth 2
```

---

### Q: Is there a limit to how many relationships I can create?

**A:** No hard limit, but practical considerations:

- **Database size:** Each relationship takes up storage
- **Performance:** Path-finding slows down with many relationships
- **Usability:** Too many links become hard to manage

**Recommendations:**
- Keep relationships meaningful (quality over quantity)
- An idea with 50+ relationships is probably over-linked
- A project with 100+ sub-tasks might need better organization
- Consider grouping related ideas into sub-projects

---

### Q: Are relationships stored in the database?

**A:** Yes, relationships are stored in the SQLite database at `~/.telos/ideas.db` (or your configured location).

They're stored separately from ideas, with:
- Unique relationship ID
- Source idea ID
- Target idea ID
- Relationship type
- Creation timestamp

---

### Q: Can I export relationships?

**A:** Currently, relationship export is not directly supported, but you can:

1. **Query the database directly:**

```bash
sqlite3 ~/.telos/ideas.db "SELECT * FROM idea_relationships;"
```

2. **Use link list for manual export:**

```bash
# List relationships for an idea
tm link list <idea-id> > relationships.txt
```

---

### Q: Do relationships affect idea scores?

**A:** No, relationships don't directly affect the Telos alignment scores.

However, you can use relationship information to make better decisions:
- Prioritize ideas that unblock other high-scoring ideas
- Focus on foundational ideas (those with many dependents)
- Review related ideas together for context

---

## Troubleshooting

### Q: Why can't I find a path between two ideas?

**A:** Several possible reasons:

**1. No path exists:**
- The ideas aren't connected through relationships
- There's a missing link in the chain

**Solution:** Create the missing relationships:
```bash
tm link create intermediate-idea target-idea depends_on
```

**2. Path is too deep:**
- Default max depth is 3
- Path might be longer than this

**Solution:** Increase max depth:
```bash
tm link path start end --max-depth 5
```

**3. Ideas are connected in wrong direction:**
- Path-finding follows relationship direction
- Your relationships might point the wrong way

**Solution:** Check relationship directions:
```bash
tm link list start-idea
tm link list end-idea
```

---

### Q: I created a relationship in the wrong direction. How do I fix it?

**A:** Remove and recreate:

```bash
# 1. List relationships to get the relationship ID
tm link list <idea-id>

# 2. Remove the wrong relationship
tm link remove rel-xyz789

# 3. Create the correct relationship
tm link create <correct-source> <correct-target> <type>
```

**Example:**
```bash
# Wrong: api depends on ui (backwards!)
tm link create api ui depends_on

# Fix it:
tm link list api             # Get rel-123456
tm link remove rel-123456
tm link create ui api depends_on  # Correct direction
```

---

### Q: Why do I get "Relationship already exists"?

**A:** You've already created this exact relationship (same source, target, and type).

**Check existing relationships:**
```bash
tm link list <idea-id>
```

**Options:**
1. **It's correct:** Nothing to do, relationship already exists
2. **Wrong type:** Remove and recreate with different type
3. **Duplicate attempt:** Just a mistake, ignore the error

---

### Q: Can I undo a relationship removal?

**A:** No, removal is permanent. You'll need to recreate it:

```bash
tm link create <source-id> <target-id> <type>
```

**Tip:** If you're unsure, don't use `--no-confirm`. The confirmation prompt helps prevent accidental deletions.

---

### Q: How do I find all ideas that depend on a specific idea?

**A:** Use `link list` and look at the "Incoming" section:

```bash
tm link list <idea-id>
```

Incoming relationships show ideas that point to this one.

**Example output:**
```
Incoming (where this idea is the target):
  1. depends_on ← [ui-abc12] Build user interface
  2. depends_on ← [api-def45] Build REST API
```

This shows that both the UI and API depend on this idea.

---

### Q: Why is path-finding slow?

**A:** Several reasons:

**1. Large relationship graph:**
- Many ideas with many relationships
- Creates a complex network to search

**2. Deep search:**
- Default max-depth is 3
- Searching deeper increases computation

**3. Highly connected graph:**
- If every idea links to many others
- Creates exponential search space

**Solutions:**

```bash
# Reduce max depth
tm link path start end --max-depth 2

# Be more selective with relationships
# Remove unnecessary links
tm link list <idea-id>
tm link remove <rel-id>
```

---

## Best Practices

### Q: How many relationships should an idea have?

**A:** There's no fixed rule, but guidelines:

**Healthy range:**
- Most ideas: 1-5 relationships
- Complex ideas: 5-15 relationships
- Major projects: 10-30 relationships

**Warning signs:**
- 50+ relationships: Probably over-linked
- 100+ relationships: Definitely too many

**Tips:**
- Link only meaningful relationships
- Focus on direct dependencies
- Use hierarchies to organize large projects
- Review and clean up periodically

---

### Q: Should I link every related idea?

**A:** No! Focus on important connections:

**Do link:**
- ✅ Hard dependencies (can't start without this)
- ✅ Clear hierarchy (this is part of that)
- ✅ Duplicates (same idea)
- ✅ Active blockers (this is preventing progress)

**Don't link:**
- ❌ Vague associations ("these might be related somehow")
- ❌ Distant connections ("these are in the same domain")
- ❌ Everything to everything

**Rule of thumb:** If you can't explain why they're linked in one sentence, don't link them.

---

### Q: How often should I review and clean up relationships?

**A:** Recommended schedule:

**Weekly:**
- Review active projects and their relationships
- Update or remove blockers that are resolved
- Check for completed dependencies

**Monthly:**
- Audit large projects for outdated relationships
- Clean up archived idea relationships
- Review duplicate markers

**Quarterly:**
- Full relationship graph review
- Remove relationships to deleted/archived ideas
- Restructure project hierarchies if needed

---

### Q: What's the best way to organize a large project?

**A:** Use hierarchical structure:

**Level 1: Main Project**
```bash
tm dump "Build E-commerce Platform"  # main-project
```

**Level 2: Major Components**
```bash
tm dump "User Management"            # component-users
tm dump "Product Catalog"            # component-products
tm dump "Shopping Cart"              # component-cart

tm link create component-users main-project part_of
tm link create component-products main-project part_of
tm link create component-cart main-project part_of
```

**Level 3: Features**
```bash
tm dump "User Registration"          # feature-register
tm dump "User Login"                 # feature-login

tm link create feature-register component-users part_of
tm link create feature-login component-users part_of
```

**Level 4: Tasks**
```bash
tm dump "Design registration form"   # task-reg-form
tm dump "Implement API endpoint"     # task-reg-api

tm link create task-reg-form feature-register part_of
tm link create task-reg-api feature-register part_of
```

This creates a clear 4-level hierarchy that's easy to navigate.

---

### Q: Should I create relationships proactively or as needed?

**A:** Mix of both:

**Create proactively:**
- When planning a project (create structure upfront)
- When you know dependencies (link before starting work)
- When organizing sprints (set up all relationships)

**Create as needed:**
- When you discover connections (link as you learn)
- When blockers arise (mark blocked_by when stuck)
- When you find duplicates (mark when reviewing)

**Don't:**
- ❌ Create all possible relationships upfront
- ❌ Wait until the project is done to create relationships

**Best approach:** Create key structural relationships (hierarchy, major dependencies) upfront, add detail relationships as you work.

---

### Q: How do I avoid circular dependencies?

**A:** Be mindful when creating `depends_on` relationships:

**Bad (circular):**
```bash
tm link create A B depends_on
tm link create B C depends_on
tm link create C A depends_on  # Creates cycle: A→B→C→A
```

**Good (linear):**
```bash
tm link create A B depends_on
tm link create B C depends_on
# C is the foundation, B builds on it, A builds on B
```

**Tips:**
- Think about logical order (what must come first?)
- Use `tm link path` to check for cycles
- If you need bidirectional dependencies, they're probably `related_to` not `depends_on`

---

## Still Have Questions?

If your question isn't answered here:

1. Check the [User Guide](../user-guide/link-command.md) for detailed documentation
2. Review the [Tutorial](../tutorials/getting-started-with-links.md) for hands-on examples
3. Browse the [Cheat Sheet](../quick-reference/link-cheatsheet.md) for quick syntax
4. Open an issue on [GitHub](https://github.com/ryacub/telos-idea-matrix/issues)

---

**FAQ Version:** 1.0
**Last Updated:** 2025-11-19
**Questions Answered:** 35+
