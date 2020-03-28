-- name: Pkg :one
SELECT * FROM packages
WHERE uuid = $1 LIMIT 1;

-- name: ModulePkgs :many
SELECT * FROM packages
WHERE module_uuid = $1;

-- name: InsertPkg :exec
INSERT INTO packages (
    uuid,
    module_uuid,
    relative_path
) VALUES (
    $1,
    $2,
    $3
) ON CONFLICT DO NOTHING;
