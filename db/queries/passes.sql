-- name: CreatePass :one
INSERT INTO passes (owner_user_id, plate_number, vehicle_brand, vehicle_color, status, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetPassByID :one
SELECT * FROM passes WHERE id = $1 AND deleted_at IS NULL;

-- name: GetPassByIDAny :one
SELECT * FROM passes WHERE id = $1;

-- name: ListPassesByOwner :many
SELECT * FROM passes
WHERE owner_user_id = $1 AND (($2::bool) OR deleted_at IS NULL)
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListPasses :many
SELECT * FROM passes
WHERE ($1::bool) OR deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: SearchPassesByPlate :many
SELECT * FROM passes
WHERE deleted_at IS NULL AND plate_number ILIKE $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdatePass :one
UPDATE passes
SET plate_number = $2,
    vehicle_brand = $3,
    vehicle_color = $4,
    status = $5,
    updated_at = now(),
    updated_by = $6
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeletePass :exec
UPDATE passes
SET deleted_at = now(),
    updated_at = now(),
    updated_by = $2
WHERE id = $1 AND deleted_at IS NULL;

-- name: RestorePass :exec
UPDATE passes
SET deleted_at = NULL,
    updated_at = now(),
    updated_by = $2
WHERE id = $1;
