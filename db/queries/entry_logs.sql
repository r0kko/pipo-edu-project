-- name: CreateEntryLog :one
INSERT INTO entry_logs (pass_id, guard_user_id, action, comment)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListEntryLogsByPass :many
SELECT * FROM entry_logs
WHERE pass_id = $1
ORDER BY action_at DESC
LIMIT $2 OFFSET $3;
