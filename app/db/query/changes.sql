-- name: DeleteChangesCommitRange :exec
DELETE FROM changes
WHERE 1=1
    AND commit_index BETWEEN sqlc.arg(commit_index_min) AND sqlc.arg(commit_index_max)
;
