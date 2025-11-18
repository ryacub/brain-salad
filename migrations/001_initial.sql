-- Create ideas table
CREATE TABLE IF NOT EXISTS ideas (
    id TEXT PRIMARY KEY,
    content TEXT NOT NULL,
    raw_score REAL,
    final_score REAL,
    patterns TEXT,  -- JSON array of pattern strings
    recommendation TEXT,
    analysis_details TEXT,  -- JSON with detailed breakdown
    created_at TEXT NOT NULL,  -- RFC3339 timestamp
    reviewed_at TEXT,
    status TEXT NOT NULL DEFAULT 'active',  -- 'active', 'archived', 'deleted'
    created_at_timestamp INTEGER GENERATED ALWAYS AS (CAST(strftime('%s', created_at) AS INTEGER)) VIRTUAL
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_ideas_status ON ideas(status);
CREATE INDEX IF NOT EXISTS idx_ideas_final_score ON ideas(final_score);
CREATE INDEX IF NOT EXISTS idx_ideas_created_at ON ideas(created_at);
CREATE INDEX IF NOT EXISTS idx_ideas_created_timestamp ON ideas(created_at_timestamp);

-- Enable foreign keys and WAL mode for better performance
PRAGMA foreign_keys = ON;
PRAGMA journal_mode = WAL;