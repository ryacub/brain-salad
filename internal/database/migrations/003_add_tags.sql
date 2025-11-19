-- 003_add_tags.sql
-- Add tags column to ideas table (idempotent)

-- Note: SQLite doesn't support IF NOT EXISTS for ALTER TABLE ADD COLUMN
-- The migration runner should handle this by checking if column exists first
-- or by catching the "duplicate column" error and ignoring it.
-- This file documents the intended migration.

ALTER TABLE ideas ADD COLUMN tags TEXT DEFAULT '[]';
