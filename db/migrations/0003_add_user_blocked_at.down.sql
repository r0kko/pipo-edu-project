DROP INDEX IF EXISTS idx_users_blocked_at;

ALTER TABLE users
    DROP COLUMN IF EXISTS blocked_at;
