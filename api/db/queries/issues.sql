-- name: StoreNewsletterIssue :one
INSERT INTO newsletter_issue (newsletter_id, user_id, sent_at) VALUES ($1, $2, $3) RETURNING id;

-- name: StoreNewsletterIssueItems :copyfrom
INSERT INTO newsletter_feed_item_status (issue_id, item_id, user_id) VALUES ($1, $2, $3);

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
SELECT ii.state, ii.item_id, i.title, i.url, i.publish_date, i.feed_id
FROM newsletter_feed_item_status AS ii
INNER JOIN feed_item AS i ON ii.item_id = i.id
WHERE issue_id = $1 AND user_id = $2;

-- name: GetIssueFeeds :many
SELECT DISTINCT f.* 
FROM feed AS f
INNER JOIN feed_item AS item ON item.feed_id = f.id
INNER JOIN newsletter_feed_item_status AS issue_item ON item.id = issue_item.item_id
WHERE issue_item.issue_id = $1 AND user_id = $2;