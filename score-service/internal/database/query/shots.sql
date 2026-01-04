-- name: CreateShot :one
WITH target_info AS (
    SELECT zones_config, max_score, has_x
    FROM target_faces tf
             JOIN qualification_rounds qr ON tf.id = qr.target_face_id
             JOIN sets s ON qr.id = s.parent_round_id
    WHERE s.id = $4 AND tf.deleted_at IS NULL AND qr.deleted_at IS NULL AND s.deleted_at IS NULL
    LIMIT 1
    )
INSERT INTO shots (
    x,
    y,
    score,
    distance_from_center,
    is_ten,
    is_x,
    is_miss,
    notes,
    set_id
) VALUES (
    $1,
    $2,
    $3,
    SQRT(POWER($1, 2) + POWER($2, 2)),
    (SELECT calculate_shot_score($1, $2, (SELECT target_face_id FROM qualification_rounds qr JOIN sets s ON qr.id = s.parent_round_id WHERE s.id = $4 LIMIT 1))) = 10,
    (SELECT calculate_shot_score($1, $2, (SELECT target_face_id FROM qualification_rounds qr JOIN sets s ON qr.id = s.parent_round_id WHERE s.id = $4 LIMIT 1))) = 10
    AND SQRT(POWER($1, 2) + POWER($2, 2)) < 30.5,
    (SELECT calculate_shot_score($1, $2, (SELECT target_face_id FROM qualification_rounds qr JOIN sets s ON qr.id = s.parent_round_id WHERE s.id = $4 LIMIT 1))) = 0,
    $4,
    $5
)
    RETURNING *;

-- name: GetShot :one
SELECT * FROM shots
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetShotsBySet :many
SELECT * FROM shots
WHERE set_id = $1 AND deleted_at IS NULL
ORDER BY created_at;

-- name: BatchCreateShots :many
INSERT INTO shots (
    x,
    y,
    score,
    distance_from_center,
    is_ten,
    is_x,
    is_miss,
    set_id
) SELECT
      unnest($1::DECIMAL[]),
      unnest($2::DECIMAL[]),
      unnest($3::INTEGER[]),
      unnest($4::DECIMAL[]),
      unnest($5::BOOLEAN[]),
      unnest($6::BOOLEAN[]),
      unnest($7::BOOLEAN[]),
      $8
          RETURNING *;