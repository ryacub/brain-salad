-- 004_sessions.sql
-- Sessions table for secure session management
-- Replaces insecure IP-based session tracking

CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    created_at TEXT NOT NULL,       -- RFC3339 format (UTC)
    expires_at TEXT NOT NULL,       -- RFC3339 format (UTC) - absolute timeout
    last_seen TEXT NOT NULL         -- RFC3339 format (UTC) - for idle timeout
);

-- Indexes for efficient queries and cleanup
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_sessions_last_seen ON sessions(last_seen);
