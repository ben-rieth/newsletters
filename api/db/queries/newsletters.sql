-- name: ListNewsletters :many
SELECT * FROM newsletters
ORDER BY created_at DESC;

-- name: GetNewsletter :one
SELECT * FROM newsletters
WHERE id = $1;

-- name: CreateNewsletter :exec
INSERT INTO newsletters (name, frequency, send_day, send_hour, send_minute)
VALUES ($1, $2, $3, $4, $5);

-- name: DeleteNewsletter :exec
DELETE FROM newsletters WHERE id = $1;