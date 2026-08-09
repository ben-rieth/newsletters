-- name: StoreNewsletterIssue :one
INSERT INTO newsletter_issue (newsletter_id, user_id, sent_at) VALUES ($1, $2, $3) RETURNING id;

-- name: StoreNewsletterIssueItems :copyfrom
INSERT INTO issue_item (issue_id, item_id, user_id, token) VALUES ($1, $2, $3, $4);

-- name: GetAllUserIssues :many
SELECT i.id, nl.id AS newsletter_id, nl.name, i.sent_at,
    CASE WHEN counts.unread_count > 0 THEN 'unread' ELSE 'read' END::item_state AS state,
    counts.item_count,
    counts.unread_count,
    preview.titles::text[] AS preview_titles
FROM newsletter_issue AS i
INNER JOIN newsletter AS nl ON i.newsletter_id = nl.id
LEFT JOIN LATERAL (
    SELECT
        COUNT(*)::int AS item_count,
        COUNT(*) FILTER (WHERE ii.state = 'unread')::int AS unread_count
    FROM issue_item AS ii
    WHERE ii.issue_id = i.id AND ii.user_id = i.user_id
) AS counts ON TRUE
LEFT JOIN LATERAL (
    SELECT COALESCE(ARRAY_AGG(top.title ORDER BY top.rank), ARRAY[]::text[]) AS titles
    FROM (
        SELECT fi.title,
            ROW_NUMBER() OVER (
                ORDER BY ii.state = 'read', fi.publish_date DESC, fi.title
            ) AS rank
        FROM issue_item AS ii
        INNER JOIN feed_item AS fi ON ii.item_id = fi.id
        WHERE ii.issue_id = i.id AND ii.user_id = i.user_id
        ORDER BY rank
        LIMIT 2
    ) AS top
) AS preview ON TRUE
WHERE i.user_id = $1
ORDER BY sent_at DESC;

-- name: GetIssue :one
SELECT i.id, nl.id AS newsletter_id, nl.name, i.sent_at,
    CASE WHEN EXISTS (
        SELECT 1 FROM issue_item AS ii WHERE ii.issue_id = i.id AND ii.state = 'unread'
    ) THEN 'unread' ELSE 'read' END::item_state AS state
FROM newsletter_issue AS i
INNER JOIN newsletter AS nl ON i.newsletter_id = nl.id
WHERE i.user_id = $1 AND i.id = $2;

-- name: GetIssueItems :many
SELECT ii.state, ii.item_id, ii.token, i.title, i.publish_date, i.feed_id
FROM issue_item AS ii
INNER JOIN feed_item AS i ON ii.item_id = i.id
WHERE issue_id = $1 AND user_id = $2
ORDER BY ii.state = 'read', i.publish_date DESC, i.title;

-- name: GetIssueFeeds :many
SELECT f.*
FROM feed AS f
INNER JOIN feed_item AS item ON item.feed_id = f.id
INNER JOIN issue_item AS issue_item ON item.id = issue_item.item_id
WHERE issue_item.issue_id = $1 AND issue_item.user_id = $2
GROUP BY f.id
ORDER BY bool_and(issue_item.state = 'read'), f.title;

-- name: GetIssueItemUrlByToken :one
SELECT fi.url FROM issue_item AS ii
INNER JOIN feed_item AS fi ON ii.item_id = fi.id
WHERE ii.token = $1;

-- name: MarkIssueItemAsReadWithToken :exec
UPDATE issue_item AS ii
SET state = 'read', updated_at = NOW()
WHERE token = $1;

-- name: DeleteItemsForIssue :exec
DELETE FROM issue_item WHERE issue_id = $1 AND user_id = $2;

-- name: DeleteIssue :exec
DELETE FROM newsletter_issue WHERE id = $1 AND user_id = $2;

-- name: DeleteAllIssueItemsForUser :exec
DELETE FROM issue_item WHERE user_id = $1;

-- name: DeleteAllIssuesForUser :exec
DELETE FROM newsletter_issue WHERE user_id = $1;

-- name: UpdateAllIssueItemsState :exec
UPDATE issue_item
SET state = $1, updated_at = NOW()
WHERE issue_id = $2 AND user_id = $3;

-- name: UpdateIssueItemState :exec
UPDATE issue_item
SET state = $1, updated_at = NOW()
WHERE item_id = $2 AND issue_id = $3 AND user_id = $4;

-- name: DoesIssueExist :one
SELECT EXISTS(SELECT 1 FROM newsletter_issue WHERE id = $1 AND user_id = $2);

-- name: DoesIssueItemExist :one
SELECT EXISTS(SELECT 1 FROM issue_item WHERE item_id = $1 AND issue_id = $2 AND user_id = $3);