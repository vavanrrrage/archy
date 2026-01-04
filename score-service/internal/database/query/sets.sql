-- name: CreateSet :one
INSERT INTO sets (
    set_number,
    max_shots,
    parent_round_id
) VALUES ($1, $2, $3)
RETURNING *;

-- name: GetSet :one
SELECT * FROM sets WHERE id = $1 AND deleted_at IS NULL;

-- name: GetSetsForQualificationRound :many
SELECT * FROM sets WHERE parent_round_id = $1 AND deleted_at IS NULL ORDER BY set_number;
