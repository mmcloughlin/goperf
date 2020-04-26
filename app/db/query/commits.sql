-- name: Commit :one
SELECT * FROM commits
WHERE sha = $1 LIMIT 1;

-- name: MostRecentCommit :one
SELECT * FROM commits
ORDER BY commit_time DESC
LIMIT 1;

-- name: MostRecentCommitWithRef :one
SELECT
    c.*
FROM
    commits AS c
    INNER JOIN commit_refs AS r
        ON c.sha=r.sha AND r.ref = sqlc.arg(ref)
ORDER BY
    c.commit_time DESC
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

-- name: BuildCommitPositions :exec
INSERT INTO commit_positions (
    SELECT
        c.sha,
        c.commit_time,
        (ROW_NUMBER() OVER (ORDER BY c.commit_time))-1 AS index
    FROM
        commits AS c
        INNER JOIN commit_refs AS r
            ON c.sha=r.sha AND r.ref = 'master'
)
ON CONFLICT (sha)
DO UPDATE SET index = EXCLUDED.index
;

-- name: MostRecentCommitIndex :one
SELECT
    MAX(index)::INT
FROM
    commit_positions
;
