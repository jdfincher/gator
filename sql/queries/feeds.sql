-- name: CreateFeed :one
INSERT INTO feeds(id, created_at, updated_at, name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: GetFeeds :many
SELECT * FROM feeds;

-- name: GetFeedID :one
SELECT id FROM feeds WHERE url = $1;

-- name: MarkFeedFetched :exec
UPDATE feeds 
SET updated_at = $1, last_fetched_at = $2
WHERE id = $3;

-- name: GetNextFeedToFetch :one
SELECT id, url, updated_at, last_fetched_at
FROM feeds
ORDER BY last_fetched_at NULLS FIRST, updated_at ASC
LIMIT 1;


