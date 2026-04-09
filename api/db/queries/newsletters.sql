-- name: ListNewsletters :many
SELECT * FROM newsletter
ORDER BY created_at DESC;

-- name: GetNewsletter :one
SELECT * FROM newsletter
WHERE id = $1;

-- name: CreateNewsletter :exec
INSERT INTO newsletter (name, frequency, send_day, send_hour, send_minute, send_timezone, next_send_time)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: DeleteNewsletter :exec
DELETE FROM newsletter WHERE id = $1;

-- name: UpdateNewsletter :exec
UPDATE newsletter SET 
    name = $1,
    frequency = $2,
    send_day = $3,
    send_hour = $4,
    send_minute = $5,
    send_timezone = $6,
    next_send_time = $7,
    updated_at = NOW()
WHERE id = $8;

-- name: DoesNewsletterExist :one
SELECT EXISTS(SELECT 1 FROM newsletter WHERE id = $1);

-- name: GetDueNewsletters :many
SELECT id, name, send_day, send_minute, send_hour, send_timezone, frequency FROM newsletter AS nl
WHERE nl.next_send_time <= NOW();