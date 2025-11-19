-- 003_add_tags.sql
-- Add tags column to ideas table

ALTER TABLE ideas ADD COLUMN tags TEXT DEFAULT '[]';
