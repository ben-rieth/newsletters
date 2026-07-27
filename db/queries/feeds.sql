-- name: GetCachedFeedDetails :one
SELECT f.id, f.title, f.url, f.html_url, f.description
FROM feed_url AS furl
INNER JOIN feed AS f ON furl.feed_id = f.id 
WHERE furl.url = $1;

-- name: DoesUserAlreadyRecieveFeed :many
SELECT nlf.alias, n.name FROM newsletter_feed AS nlf
INNER JOIN newsletter AS n ON nlf.newsletter_id = n.id
WHERE nlf.user_id = $1 AND nlf.feed_id = $2;

-- name: SaveFeedDetails :one
INSERT INTO feed (title, url, html_url, description, last_retrieved_at) 
VALUES ($1, $2, $3, $4, $5)
RETURNING id;

-- name: UpdateCachedFeed :exec
UPDATE feed 
SET title = $1, html_url = $2, description = $3, last_retrieved_at = $4, updated_at = NOW()
WHERE id = $5;

-- name: SaveFeedItemDetails :copyfrom
INSERT INTO feed_item (feed_id, title, url, publish_date, retrieved_at) VALUES ($1, $2, $3, $4, $5);

-- name: SaveFeedUrls :copyfrom
INSERT INTO feed_url (feed_id, url, source) VALUES ($1, $2, $3);

-- name: GetFeedsForNewsletter :many
SELECT
    f.title, f.description, f.url, f.html_url, nf.alias, nf.id,
    f.disabled_until, f.last_retrieved_at
FROM newsletter_feed AS nf
INNER JOIN feed AS f ON nf.feed_id = f.id
WHERE newsletter_id = $1;

-- name: AddNewsletterFeed :exec
INSERT INTO newsletter_feed (newsletter_id, feed_id, user_id, alias) VALUES ($1, $2, $3, $4);

-- name: DoesFeedExist :one
SELECT EXISTS(
    SELECT 1 FROM newsletter_feed AS f 
    INNER JOIN newsletter AS nl ON f.newsletter_id = nl.id 
    WHERE f.id = $1 AND f.newsletter_id = $2 AND nl.user_id = $3
);

-- name: UpdateNewsletterFeed :exec
UPDATE newsletter_feed 
SET alias = $1, updated_at = NOW() 
WHERE newsletter_id = $2 AND id = $3 AND user_id = $4;

-- name: DeleteNewsletterFeed :exec
DELETE FROM newsletter_feed WHERE newsletter_id = $1 AND id = $2 AND user_id = $3;

-- name: GetFeedsForManyNewsletters :many
SELECT
    f.id AS global_feed_id, nlf.id AS newsletter_feed_id,
    f.title, f.url, f.html_url, f.last_retrieved_at,
    nlf.newsletter_id, nlf.alias
FROM newsletter_feed AS nlf
INNER JOIN feed AS f ON nlf.feed_id = f.id
WHERE nlf.newsletter_id = ANY($1::UUID[]);

-- name: DeleteAllFeedsInNewsletter :exec
DELETE FROM newsletter_feed WHERE newsletter_id = $1;

-- name: DeleteAllNewsletterFeedsForUser :exec
DELETE FROM newsletter_feed WHERE user_id = $1;

-- name: DeleteNewsletterFeedItemStatuses :exec
DELETE FROM issue_item WHERE user_id = $1;

-- name: GetFeedItemsPublishedAfter :many
SELECT item.id, title, url, publish_date FROM feed_item AS item
WHERE feed_id = @global_feed_id AND publish_date > @publish_date_greater_than
AND NOT EXISTS (
    SELECT 1 FROM newsletter_feed_filter AS ff
    WHERE ff.newsletter_feed_id = @newsletter_feed_id AND ff.user_id = @user_id
        AND (
            (ff.operator = 'contains' AND (
                ('title' = ff.field AND item.title ILIKE '%' || ff.pattern || '%')
                OR
                ('url' = ff.field AND item.url ILIKE '%' || ff.pattern || '%')
            ))
            OR
            (ff.operator = 'does_not_contain' AND (
                ('title' = ff.field AND item.title NOT ILIKE '%' || ff.pattern || '%')
                OR
                ('url' = ff.field AND item.url NOT ILIKE '%' || ff.pattern || '%')
            ))
        )
);

-- name: PreviewFeed :many
SELECT item.id AS item_id, item.title, item.url, ff.id AS filter_id, ff.field, ff.operator, ff.pattern
FROM newsletter_feed AS nlf
INNER JOIN feed AS f ON nlf.feed_id = f.id
INNER JOIN feed_item AS item ON item.feed_id = f.id
LEFT JOIN newsletter_feed_filter AS ff ON ff.user_id = @user_id AND ff.newsletter_feed_id = @newsletter_feed_id
AND (
    (ff.operator = 'contains' AND (
        ('title' = ff.field AND item.title ILIKE '%' || ff.pattern || '%')
        OR
        ('url' = ff.field AND item.url ILIKE '%' || ff.pattern || '%')
    ))
    OR
    (ff.operator = 'does_not_contain' AND (
        ('title' = ff.field AND item.title NOT ILIKE '%' || ff.pattern || '%')
        OR
        ('url' = ff.field AND item.url NOT ILIKE '%' || ff.pattern || '%')
    ))
)
WHERE nlf.newsletter_id = @newsletter_id 
    AND nlf.id = @newsletter_feed_id 
    AND item.publish_date > @publish_date_greater_than;

-- name: UpdateFeedLastRetrievedTime :exec
UPDATE feed SET last_retrieved_at = $1, updated_at = NOW() WHERE id = $2;

-- name: GetFeedById :one
SELECT
    nlf.id, f.title, f.description, f.url, f.html_url, nlf.alias,
    f.disabled_until, f.last_retrieved_at
FROM newsletter_feed AS nlf
INNER JOIN feed AS f ON nlf.feed_id = f.id
WHERE nlf.id = $1 AND nlf.user_id = $2;

-- name: GetLastFeedFailure :one
SELECT ff.occurred_at, ff.message FROM newsletter_feed AS nlf
INNER JOIN feed_fetch_failure AS ff ON ff.feed_id = nlf.feed_id
WHERE nlf.id = $1 AND nlf.user_id = $2 AND ff.occurred_at > $3
ORDER BY ff.occurred_at DESC
LIMIT 1;

-- name: GetLastFeedFailuresForNewsletter :many
SELECT DISTINCT ON (nlf.id)
    nlf.id AS newsletter_feed_id, ff.occurred_at, ff.message
FROM newsletter_feed AS nlf
INNER JOIN feed_fetch_failure AS ff ON ff.feed_id = nlf.feed_id
WHERE nlf.newsletter_id = $1 AND nlf.user_id = $2 AND ff.occurred_at > $3
ORDER BY nlf.id, ff.occurred_at DESC;

-- name: RecordFeedFetchFailure :exec
INSERT INTO feed_fetch_failure (feed_id, url, kind, status_code, message)
VALUES ($1, $2, $3, $4, $5);

-- name: DeleteFeedFetchFailuresBefore :execrows
DELETE FROM feed_fetch_failure WHERE occurred_at < $1;

-- name: CountRecentFeedFailures :one
SELECT COUNT(*) FROM feed_fetch_failure
WHERE feed_id = $1 AND occurred_at > $2;

-- name: GetFeedBreakerState :one
SELECT disabled_until, disable_count, last_retrieved_at FROM feed WHERE id = $1;

-- name: DisableFeedUntil :exec
UPDATE feed
SET disabled_until = $1, disable_count = disable_count + 1, updated_at = NOW()
WHERE id = $2 AND (disabled_until IS NULL OR disabled_until <= NOW());

-- name: ResetFeedBreaker :exec
UPDATE feed SET disabled_until = NULL, disable_count = 0, updated_at = NOW()
WHERE id = $1 AND (disabled_until IS NOT NULL OR disable_count > 0);
