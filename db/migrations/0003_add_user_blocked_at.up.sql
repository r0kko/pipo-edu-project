ALTER TABLE users
    ADD COLUMN IF NOT EXISTS blocked_at TIMESTAMPTZ NULL;

CREATE INDEX IF NOT EXISTS idx_users_blocked_at ON users (blocked_at);
