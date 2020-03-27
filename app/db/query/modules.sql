-- name: Module :one
SELECT * FROM modules
WHERE uuid = $1 LIMIT 1;

-- name: InsertModule :exec
INSERT INTO modules (
    uuid,
    path,
    version
) VALUES (
    $1,
    $2,
    $3
) ON CONFLICT DO NOTHING;
