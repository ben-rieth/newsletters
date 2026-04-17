-- name: AddFeedFilter :exec
INSERT INTO newsletter_feed_filter (newsletter_feed_id, user_id, field, operator, pattern)
VALUES ($1, $2, $3, $4, $5);

-- name: GetFiltersForManyFeeds :many
SELECT id, field, operator, pattern, newsletter_feed_id FROM newsletter_feed_filter
WHERE user_id = $1 AND newsletter_feed_id = ANY(@feed_ids::UUID[]);

-- name: GetFiltersForFeed :many
SELECT id, field, operator, pattern FROM newsletter_feed_filter
WHERE user_id = $1 AND newsletter_feed_id = $2;

-- name: DoesFilterExist :one
SELECT EXISTS(
    SELECT 1 FROM newsletter_feed_filter AS fil
    INNER JOIN newsletter_feed AS nlf ON fil.newsletter_feed_id = nlf.id
    WHERE fil.id = @filter_id AND fil.user_id = $1 AND nlf.newsletter_id = $2 AND nlf.id = @newsletter_feed_id
);

-- name: UpdateFeedFilter :exec
UPDATE newsletter_feed_filter 
SET field = $1, operator = $2, pattern = $3, updated_at = NOW()
WHERE id = $4;

-- name: DeleteFeedFilter :exec
DELETE FROM newsletter_feed_filter WHERE id = $1 AND user_id = $2;
