-- name: CreateGuestRequest :one
INSERT INTO guest_requests (resident_user_id, guest_full_name, plate_number, valid_from, valid_to, status, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetGuestRequestByID :one
SELECT * FROM guest_requests WHERE id = $1 AND deleted_at IS NULL;

-- name: GetGuestRequestByIDAny :one
SELECT * FROM guest_requests WHERE id = $1;

-- name: ListGuestRequestsByResident :many
SELECT * FROM guest_requests
WHERE resident_user_id = $1 AND (($2::bool) OR deleted_at IS NULL)
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListGuestRequests :many
SELECT * FROM guest_requests
WHERE ($1::bool) OR deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateGuestRequest :one
UPDATE guest_requests
SET guest_full_name = $2,
    plate_number = $3,
    valid_from = $4,
    valid_to = $5,
    status = $6,
    updated_at = now(),
    updated_by = $7
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteGuestRequest :exec
UPDATE guest_requests
SET deleted_at = now(),
    updated_at = now(),
    updated_by = $2
WHERE id = $1 AND deleted_at IS NULL;

-- name: RestoreGuestRequest :exec
UPDATE guest_requests
SET deleted_at = NULL,
    updated_at = now(),
    updated_by = $2
WHERE id = $1;
