-- name: StoreNewsletterIssue :one
INSERT INTO newsletter_issue (newsletter_id, user_id, sent_at) VALUES ($1, $2, $3) RETURNING id;

-- name: StoreNewsletterIssueItems :copyfrom
INSERT INTO newsletter_feed_item_status (issue_id, item_id, user_id) VALUES ($1, $2, $3);
