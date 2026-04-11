-- name: AddFeed :exec
INSERT INTO feed (name, url, newsletter_id)
VALUES ($1, $2, $3);

-- name: GetFeedsForNewsletter :many
SELECT * FROM feed WHERE newsletter_id = $1;

-- name: DoesFeedExist :one
SELECT EXISTS(
    SELECT 1 FROM feed AS f 
    INNER JOIN newsletter AS nl ON f.newsletter_id = nl.id 
    WHERE f.id = $1 AND f.newsletter_id = $2 AND nl.user_id = $3
);

-- name: UpdateFeed :exec
UPDATE feed SET name = $1, url = $2 WHERE newsletter_id = $3 AND id = $4;

-- name: DeleteFeed :exec
DELETE FROM feed WHERE newsletter_id = $1 AND id = $2;

-- name: DeleteAllFeedsInNewsletter :exec
DELETE FROM feed WHERE newsletter_id = $1;

-- name: GetFeedsForManyNewsletters :many
SELECT newsletter_id, id, name, url FROM feed WHERE newsletter_id = ANY($1::string[]);