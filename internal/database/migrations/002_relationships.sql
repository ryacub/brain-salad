-- 002_relationships.sql
-- Add idea_relationships table for linking ideas

-- Idea relationships table: stores connections between ideas
CREATE TABLE IF NOT EXISTS idea_relationships (
    id TEXT PRIMARY KEY,
    source_idea_id TEXT NOT NULL,
    target_idea_id TEXT NOT NULL,
    relationship_type TEXT NOT NULL,
    created_at TEXT NOT NULL,       -- RFC3339 format (UTC)
    FOREIGN KEY (source_idea_id) REFERENCES ideas (id) ON DELETE CASCADE,
    FOREIGN KEY (target_idea_id) REFERENCES ideas (id) ON DELETE CASCADE,
    UNIQUE(source_idea_id, target_idea_id, relationship_type)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_relationships_source
    ON idea_relationships(source_idea_id);

CREATE INDEX IF NOT EXISTS idx_relationships_target
    ON idea_relationships(target_idea_id);

CREATE INDEX IF NOT EXISTS idx_relationships_type
    ON idea_relationships(relationship_type);

-- Composite index for common query pattern
CREATE INDEX IF NOT EXISTS idx_relationships_source_type
    ON idea_relationships(source_idea_id, relationship_type);
