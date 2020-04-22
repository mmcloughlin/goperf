-- name: Commit :one
SELECT * FROM commits
WHERE sha = $1 LIMIT 1;

-- name: MostRecentCommit :one
SELECT * FROM commits
ORDER BY commit_time DESC
LIMIT 1;

-- name: InsertCommit :exec
INSERT INTO commits (
    sha,
    tree,
    parents,
    author_name,
    author_email,
    author_time,
    committer_name,
    committer_email,
    commit_time,
    message
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10
) ON CONFLICT DO NOTHING;

-- name: InsertCommitRef :exec
INSERT INTO commit_refs (
    sha,
    ref
) VALUES (
    $1,
    $2
) ON CONFLICT DO NOTHING;
