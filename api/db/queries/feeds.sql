-- name: GetCachedFeedDetails :one
SELECT f.id, f.title, f.url, f.description 
FROM feed_url AS furl
INNER JOIN feed AS f ON furl.feed_id = f.id 
WHERE furl.url = $1;

-- name: SaveFeedDetails :one
INSERT INTO feed (title, url, description, last_retrieved_at) 
VALUES ($1, $2, $3, $4)
RETURNING id;

-- name: SaveFeedItemDetails :copyfrom
INSERT INTO feed_item (feed_id, title, url, publish_date, retrieved_at) VALUES ($1, $2, $3, $4, $5);

-- name: SaveFeedUrls :copyfrom
INSERT INTO feed_url (feed_id, url, source) VALUES ($1, $2, $3);

-- name: GetFeedsForNewsletter :many
SELECT f.title, f.description, f.url, nf.alias FROM newsletter_feed AS nf
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
UPDATE newsletter_feed SET alias = $1 WHERE newsletter_id = $2 AND id = $3;

-- name: DeleteNewsletterFeed :exec
DELETE FROM newsletter_feed WHERE newsletter_id = $1 AND id = $2;

-- name: GetFeedsForManyNewsletters :many
SELECT f.id, f.title, f.url, f.last_retrieved_at, nlf.newsletter_id FROM newsletter_feed AS nlf
INNER JOIN feed AS f ON nlf.feed_id = f.id
WHERE nlf.newsletter_id = ANY($1::UUID[]);

-- name: DeleteAllFeedsInNewsletter :exec
DELETE FROM newsletter_feed WHERE newsletter_id = $1;

-- name: DeleteAllNewsletterFeedsForUser :exec
DELETE FROM newsletter_feed WHERE user_id = $1;

-- name: DeleteNewsletterFeedItemStatuses :exec
DELETE FROM newsletter_feed_item_status WHERE user_id = $1;