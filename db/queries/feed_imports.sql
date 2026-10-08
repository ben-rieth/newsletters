-- name: CreateFeedImports :copyfrom
INSERT INTO newsletter_feed_import (newsletter_id, user_id, url, alias, status, filters)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetPendingFeedImportIds :many
SELECT id FROM newsletter_feed_import WHERE state = 'pending';

-- name: GetPendingFeedImportIdsForNewsletters :many
SELECT id FROM newsletter_feed_import
WHERE state = 'pending' AND newsletter_id = ANY(@newsletter_ids::UUID[]);

-- name: GetFeedImportsForNewsletter :many
SELECT id, url, alias, state, error FROM newsletter_feed_import
WHERE newsletter_id = $1 AND user_id = $2
ORDER BY created_at, url;

-- name: GetPendingFeedImport :one
SELECT * FROM newsletter_feed_import WHERE id = $1 AND state = 'pending';

-- name: MarkFeedImportFailed :exec
UPDATE newsletter_feed_import
SET state = 'failed', error = $2, updated_at = NOW()
WHERE id = $1;

-- name: RetryFeedImport :execrows
UPDATE newsletter_feed_import
SET state = 'pending', error = '', updated_at = NOW()
WHERE id = $1 AND newsletter_id = $2 AND user_id = $3 AND state = 'failed';

-- name: DeleteFeedImport :execrows
DELETE FROM newsletter_feed_import WHERE id = $1 AND newsletter_id = $2 AND user_id = $3;

-- name: DeleteResolvedFeedImport :execrows
DELETE FROM newsletter_feed_import WHERE id = $1;

-- name: DeleteFeedImportsForNewsletter :exec
DELETE FROM newsletter_feed_import WHERE newsletter_id = $1 AND user_id = $2;

-- name: DeleteFeedImportsForUser :exec
DELETE FROM newsletter_feed_import WHERE user_id = $1;

-- name: CountFeedImportsByNewsletter :many
SELECT
    newsletter_id,
    COUNT(*) FILTER (WHERE state = 'pending')::INT AS pending,
    COUNT(*) FILTER (WHERE state = 'failed')::INT AS failed
FROM newsletter_feed_import
WHERE user_id = $1
GROUP BY newsletter_id;
