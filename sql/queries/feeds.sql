-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
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
SELECT feeds.name, feeds.url, feeds.user_id FROM feeds;

-- name: GetFeedByUrl :one
SELECT * FROM feeds WHERE feeds.url=$1;

-- name: DeleteFeeds :exec
DELETE from feeds;



-- name: CreateFeedFollow :one
WITH inserted_feed_follow as (
    INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
    VALUES (
        $1,
        $2,
        $3,
        $4,
        $5
    )
    RETURNING *
)
SELECT  inserted_feed_follow.*,
        feeds.name AS feed_name,
        users.name AS user_name
FROM inserted_feed_follow
INNER JOIN users ON users.id = inserted_feed_follow.user_id
INNER JOIN feeds ON feeds.id = inserted_feed_follow.feed_id;

-- name: GetFeedFollowsForUser :many
SELECT  feed_follows.*,
        feeds.name AS feed_name,
        users.name AS user_name
 FROM feed_follows
 INNER JOIN feeds ON feed_follows.feed_id = feeds.id
 INNER JOIN users ON feed_follows.user_id = users.id
 WHERE feed_follows.user_id = $1;

-- name: MarkFeedFetched :one
UPDATE feeds
SET last_fetched_at = NOW(),
updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteFeedsFollow :exec
WITH selected_feed_id as (
    SELECT feeds.id from feeds WHERE feeds.url = $1
)
 DELETE FROM feed_follows WHERE feed_follows.user_id = $2 AND feed_follows.feed_id = (SELECT id from selected_feed_id);

-- name: GetNextFeedToFetch :one
SELECT * FROM feeds ORDER BY last_fetched_at ASC NULLS FIRST LIMIT 1;

-- name: GetFeedById :one
SELECT * FROM feeds WHERE feeds.id = $1;