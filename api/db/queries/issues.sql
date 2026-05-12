-- name: StoreNewsletterIssue :one
INSERT INTO newsletter_issue (newsletter_id, user_id, sent_at) VALUES ($1, $2, $3) RETURNING id;

-- name: StoreNewsletterIssueItems :copyfrom
INSERT INTO issue_item (issue_id, item_id, user_id, token) VALUES ($1, $2, $3, $4);

-- name: GetAllUserIssues :many
SELECT i.id, nl.id, nl.name, i.sent_at FROM newsletter_issue AS i 
INNER JOIN newsletter AS nl ON i.newsletter_id = nl.id
WHERE i.user_id = $1
ORDER BY sent_at DESC;

-- name: GetIssue :one
SELECT i.id, nl.name, i.sent_at FROM newsletter_issue AS i 
INNER JOIN newsletter AS nl ON i.newsletter_id = nl.id
WHERE i.user_id = $1 AND i.id = $2;

-- name: GetIssueItems :many
SELECT ii.state, ii.item_id, ii.token, i.title, i.publish_date, i.feed_id
FROM issue_item AS ii
INNER JOIN feed_item AS i ON ii.item_id = i.id
WHERE issue_id = $1 AND user_id = $2;

-- name: GetIssueFeeds :many
SELECT DISTINCT f.* 
FROM feed AS f
INNER JOIN feed_item AS item ON item.feed_id = f.id
INNER JOIN issue_item AS issue_item ON item.id = issue_item.item_id
WHERE issue_item.issue_id = $1 AND user_id = $2;

-- name: GetIssueItemUrlByToken :one
SELECT fi.url FROM issue_item AS ii
INNER JOIN feed_item AS fi ON ii.item_id = fi.id
WHERE ii.token = $1;

-- name: MarkIssueItemAsRead :exec
UPDATE issue_item AS ii
SET state = 'read'
WHERE token = $1;

-- name: DeleteItemsForIssue :exec
DELETE FROM issue_item WHERE issue_id = $1 AND user_id = $2;

-- name: DeleteIssue :exec
DELETE FROM newsletter_issue WHERE id = $1 AND user_id = $2;