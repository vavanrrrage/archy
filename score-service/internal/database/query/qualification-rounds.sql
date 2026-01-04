-- name: CreateQualificationRound :one
INSERT INTO qualification_rounds (
    external_user_id,
    round_type,
    name,
    distance,
    total_sets,
    shots_per_set,
    target_face_id,
    notes,
    start_time
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    RETURNING *;

-- name: GetQualificationRound :one
SELECT * FROM qualification_rounds WHERE id = $1 AND deleted_at IS NULL;

-- name: GetQualificationRoundsForUser :many
SELECT * FROM qualification_rounds WHERE external_user_id = $1 AND deleted_at IS NULL;