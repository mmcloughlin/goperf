-- name: Commit :one
SELECT * FROM commits
WHERE sha = $1 LIMIT 1;
