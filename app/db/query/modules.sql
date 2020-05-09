-- name: Module :one
SELECT * FROM modules
WHERE uuid = $1 LIMIT 1;

-- name: Modules :many
SELECT
    *
FROM
    modules
ORDER BY
    CASE
        WHEN path='std' THEN '0' || path
        WHEN path LIKE 'golang.org/x/%' THEN '1' || path
        ELSE '2' || path
    END
;

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
