-- name: Properties :one
SELECT * FROM properties
WHERE uuid = $1 LIMIT 1;

-- name: InsertProperties :exec
INSERT INTO properties (
    uuid,
    fields
) VALUES (
    $1,
    $2
) ON CONFLICT DO NOTHING;
