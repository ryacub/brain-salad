# Tutorial: Getting Started with Idea Relationships

**Time:** 10-15 minutes
**Level:** Beginner
**Prerequisites:** Telos Matrix installed, at least one idea in your database

## What You'll Learn

By the end of this tutorial, you'll know how to:

- Create your first relationship between ideas
- View relationships for an idea
- Find connections between ideas
- Organize a small project with links
- Use relationships to track dependencies

## Before You Start

Make sure you have the Telos Matrix CLI installed and can run commands:

```bash
tm --help
```

If you don't have any ideas yet, create a few test ideas:

```bash
tm dump "Set up project infrastructure"
tm dump "Design database schema"
tm dump "Build REST API"
tm dump "Create user interface"
```

## Step 1: Create Your First Relationship

Let's start with a simple dependency. First, let's see what ideas we have:

```bash
tm review pending --limit 5
```

You should see a list of your ideas with their IDs. They'll look something like:

```
1. [abc12345] Set up project infrastructure
2. [def67890] Design database schema
3. [ghi13579] Build REST API
4. [jkl24680] Create user interface
```

**Note:** You can use the shortened ID (first 8 characters) in most commands.

Now, let's create a relationship. The API depends on the database schema, so let's link them:

```bash
# Replace with your actual IDs
tm link create ghi13579 def67890 depends_on
```

The command will show you a preview:

```
Creating relationship:
  Source: [ghi13579] Build REST API
  Target: [def67890] Design database schema
  Type: depends_on

Continue? (y/n):
```

Type `y` and press Enter.

You should see:

```
✓ Relationship created successfully (ID: rel-xxxxxxxx)
```

**What just happened:**

You told Telos Matrix that "Build REST API" depends on "Design database schema". This means you can't start building the API until the database schema is designed.

## Step 2: View Your Relationships

Now let's see the relationships for the API idea:

```bash
tm link list ghi13579
```

You should see output like:

```
🔗 Relationships for idea: [ghi13579]
   Build REST API

Outgoing (where this idea is the source):
  1. depends_on → [def67890] Design database schema
     ID: rel-abc123 | Created: 2025-11-19 10:30

Total: 1 relationship
```

**Understanding the output:**

- **Outgoing relationships:** This idea points to other ideas
- **Incoming relationships:** Other ideas point to this idea (we'll see this next)
- **Relationship ID:** Use this to remove the relationship later

## Step 3: Create a Chain of Dependencies

Let's add more relationships to build a dependency chain:

```bash
# Database depends on infrastructure
tm link create def67890 abc12345 depends_on

# UI depends on API
tm link create jkl24680 ghi13579 depends_on
```

Now check the API idea again:

```bash
tm link list ghi13579
```

You should now see:

```
🔗 Relationships for idea: [ghi13579]
   Build REST API

Outgoing (where this idea is the source):
  1. depends_on → [def67890] Design database schema
     ID: rel-abc123 | Created: 2025-11-19 10:30

Incoming (where this idea is the target):
  1. depends_on ← [jkl24680] Create user interface
     ID: rel-def456 | Created: 2025-11-19 10:35

Total: 2 relationships
```

**What this means:**

- The API **depends on** the database schema (outgoing)
- The UI **depends on** the API (incoming)

## Step 4: Find Paths Between Ideas

Now we have a chain: UI → API → Database → Infrastructure

Let's verify this with the path-finding command:

```bash
tm link path jkl24680 abc12345
```

You should see:

```
🔍 Finding paths from [jkl24680] to [abc12345]...

Path 1 (3 hops):
  [jkl24680] Create user interface
    → depends_on →
  [ghi13579] Build REST API
    → depends_on →
  [def67890] Design database schema
    → depends_on →
  [abc12345] Set up project infrastructure

Found 1 path(s)
```

**What this shows:**

The path command found the dependency chain from the UI all the way back to the infrastructure. This helps you understand the order in which work needs to be done.

## Step 5: Organize Ideas into a Project

Now let's create a project and make these tasks part of it:

```bash
# Create a project idea
tm dump "Build task management app"
```

Note the ID of the project (let's say it's `proj12345`).

Now link the tasks to the project using `part_of`:

```bash
tm link create abc12345 proj12345 part_of
tm link create def67890 proj12345 part_of
tm link create ghi13579 proj12345 part_of
tm link create jkl24680 proj12345 part_of
```

Now view the project structure:

```bash
tm link show proj12345 --type part_of
```

You should see all the tasks that are part of this project:

```
🔗 Related ideas for: [proj12345]
   Build task management app

1. Set up project infrastructure
   ID: abc12345 | Status: pending 📊 7.5/10
   Created: 2025-11-19 09:00

2. Design database schema
   ID: def67890 | Status: pending 📊 8.0/10
   Created: 2025-11-19 09:15

3. Build REST API
   ID: ghi13579 | Status: pending 📊 7.8/10
   Created: 2025-11-19 09:30

4. Create user interface
   ID: jkl24680 | Status: pending 📊 7.2/10
   Created: 2025-11-19 09:45

Found 4 related ideas
```

## Step 6: Add a Blocker

Let's say you discovered that deployment is blocked by something. First, create the blocker idea:

```bash
tm dump "Waiting for AWS account approval"
```

Let's say the deployment task ID is `deploy99`.

Now mark it as blocked:

```bash
tm link create deploy99 aws-approval blocked_by
```

## Step 7: Mark Duplicate Ideas

Sometimes you capture the same idea twice. Let's create a duplicate scenario:

```bash
tm dump "Add user authentication"
tm dump "Implement login system"
```

These are essentially the same idea. Mark them as duplicates:

```bash
# Replace with your actual IDs
tm link create auth-001 auth-002 duplicate
```

Now you can archive one of them:

```bash
tm update auth-002 --status archived
```

## Step 8: Remove a Relationship

If you created a relationship by mistake, you can remove it.

First, find the relationship ID:

```bash
tm link list <idea-id>
```

Note the relationship ID from the output (e.g., `rel-abc123`).

Then remove it:

```bash
tm link remove rel-abc123
```

You'll see a confirmation:

```
Removing relationship:
  ID: rel-abc123
  [api-ghi13] Build REST API
    depends_on →
  [db-def67] Design database schema

Are you sure? (y/n):
```

Type `y` to confirm, or `n` to cancel.

## Practical Exercise

Now try this on your own:

**Scenario:** You're planning a blog website.

1. Create these ideas:
   - "Launch personal tech blog"
   - "Choose blogging platform"
   - "Design blog theme"
   - "Write first 5 blog posts"
   - "Set up custom domain"
   - "Configure SEO"

2. Create a project structure:
   - Make "Launch personal tech blog" the main project
   - Link all other ideas to it using `part_of`

3. Add dependencies:
   - Theme design depends on platform choice
   - Writing posts depends on theme being ready
   - SEO depends on custom domain
   - Launching depends on everything else

4. View the project:
   - Use `tm link show` to see all project tasks
   - Use `tm link path` to find the critical path

**Solution:**

```bash
# Create ideas
tm dump "Launch personal tech blog"        # proj-blog
tm dump "Choose blogging platform"         # task-platform
tm dump "Design blog theme"                # task-theme
tm dump "Write first 5 blog posts"         # task-posts
tm dump "Set up custom domain"             # task-domain
tm dump "Configure SEO"                    # task-seo

# Link to project
tm link create task-platform proj-blog part_of
tm link create task-theme proj-blog part_of
tm link create task-posts proj-blog part_of
tm link create task-domain proj-blog part_of
tm link create task-seo proj-blog part_of

# Add dependencies
tm link create task-theme task-platform depends_on
tm link create task-posts task-theme depends_on
tm link create task-seo task-domain depends_on

# View project
tm link show proj-blog --type part_of

# Find critical path
tm link path task-posts task-platform
```

## What You've Learned

Congratulations! You now know how to:

- ✅ Create relationships between ideas using `link create`
- ✅ View all relationships for an idea using `link list`
- ✅ Show related ideas with details using `link show`
- ✅ Find dependency paths using `link path`
- ✅ Organize ideas into projects using `part_of`
- ✅ Track dependencies using `depends_on`
- ✅ Mark blockers using `blocked_by`
- ✅ Identify duplicates using `duplicate`
- ✅ Remove relationships using `link remove`

## Next Steps

Now that you understand the basics, you can:

1. **Explore more relationship types**
   - Try `related_to` for loose associations
   - Use `similar_to` for comparing approaches
   - Experiment with `parent` and `child` for hierarchies

2. **Read the comprehensive guide**
   - Check out the [Link Command User Guide](../user-guide/link-command.md) for detailed examples and best practices

3. **Learn advanced workflows**
   - Sprint planning with dependencies
   - Research project organization
   - Complex project hierarchies

4. **Get quick answers**
   - Bookmark the [Cheat Sheet](../quick-reference/link-cheatsheet.md)
   - Browse the [FAQ](../faq/link-command-faq.md)

## Common Questions

**Q: Can I use shortened IDs?**

A: Yes! You only need the first 8 characters of an ID. Instead of `abc12345-def6-7890-ghi1-jkl234567890`, just use `abc12345`.

**Q: What if I create a relationship in the wrong direction?**

A: Remove it with `tm link remove` and recreate it the right way.

**Q: Do I need to create both directions for relationships?**

A: No! For symmetric types like `related_to`, `similar_to`, and `duplicate`, the relationship automatically works both ways.

**Q: Can I link more than two ideas?**

A: Each relationship connects exactly two ideas, but an idea can have many relationships. For example, a project can have 10 tasks all using `part_of`.

**Q: What's the difference between `depends_on` and `blocked_by`?**

A: They're similar, but:
- `depends_on` indicates a prerequisite (B must complete before A can start)
- `blocked_by` indicates an active blocker (B is preventing A from proceeding)

Use whichever makes more sense for your workflow.

## Troubleshooting

**Problem:** "Idea not found"

**Solution:** Check the ID with `tm review pending` and copy the exact ID.

---

**Problem:** "Invalid relationship type"

**Solution:** Make sure you're using underscores (`depends_on`, not `depends-on`) and lowercase.

---

**Problem:** Can't find the relationship ID

**Solution:** Use `tm link list <idea-id>` to see all relationships and their IDs.

---

## Conclusion

You've completed the Getting Started with Links tutorial! You now have the foundational knowledge to organize your ideas, track dependencies, and build project structures in Telos Matrix.

Start small, experiment with different relationship types, and gradually build more complex idea networks as you get comfortable with the system.

Happy linking!

---

**Tutorial Version:** 1.0
**Last Updated:** 2025-11-19
**Difficulty:** Beginner
**Estimated Time:** 10-15 minutes
