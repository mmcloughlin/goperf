-- name: DataFile :one
SELECT * FROM datafiles
WHERE uuid = $1 LIMIT 1;

-- name: InsertDataFile :exec
INSERT INTO datafiles (
    uuid,
    name,
    sha256
) VALUES (
    $1,
    $2,
    $3
) ON CONFLICT DO NOTHING;
