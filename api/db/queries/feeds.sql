-- name: AddFeed :exec
INSERT INTO feed (name, url, newsletter_id)
VALUES ($1, $2, $3);

-- name: GetFeedsForNewsletter :many
SELECT * FROM feed WHERE newsletter_id = $1;