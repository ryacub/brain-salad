-- 001_initial.sql
-- Initial database schema for Telos Idea Matrix

-- Ideas table: stores captured ideas with analysis
CREATE TABLE IF NOT EXISTS ideas (
    id TEXT PRIMARY KEY,
    content TEXT NOT NULL,
    raw_score REAL,
    final_score REAL,
    patterns TEXT,                  -- JSON-serialized []string
    recommendation TEXT,
    analysis_details TEXT,          -- JSON-serialized Analysis struct
    created_at TEXT NOT NULL,       -- RFC3339 format (UTC)
    reviewed_at TEXT,               -- RFC3339 format (UTC)
    status TEXT NOT NULL DEFAULT 'active'
);

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_ideas_created_at ON ideas(created_at);
CREATE INDEX IF NOT EXISTS idx_ideas_final_score ON ideas(final_score);
CREATE INDEX IF NOT EXISTS idx_ideas_status ON ideas(status);
CREATE INDEX IF NOT EXISTS idx_ideas_status_score ON ideas(status, final_score);
