-- name: ListNewsletters :many
SELECT * FROM newsletter
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetNewsletter :one
SELECT * FROM newsletter
WHERE id = $1 AND user_id = $2;

-- name: CreateNewsletter :exec
INSERT INTO newsletter (name, frequency, send_day, send_hour, send_minute, send_timezone, next_send_time, last_sent_at, user_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: DeleteNewsletter :exec
DELETE FROM newsletter WHERE id = $1 AND user_id = $2;

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
WHERE id = $8 AND user_id = $9;

-- name: DoesNewsletterExist :one
SELECT EXISTS(SELECT 1 FROM newsletter WHERE id = $1 AND user_id = $2);

-- name: GetDueNewsletters :many
SELECT nl.id, name, send_day, send_minute, send_hour, send_timezone, frequency, u.email, u.id AS user_id, last_sent_at, nl.unsubscribe_token
FROM newsletter AS nl
INNER JOIN app_user AS u ON nl.user_id = u.id
WHERE nl.next_send_time <= NOW() AND nl.status = 'active';

-- name: UpdateNewsletterSendTime :exec
UPDATE newsletter SET next_send_time = $1, updated_at = NOW() WHERE id = $2 AND user_id = $3;

-- name: UpdateNewsletterSendTimes :exec
UPDATE newsletter SET next_send_time = $1, last_sent_at = $2, updated_at = NOW() WHERE id = $3 AND user_id = $4;

-- name: DeleteAllNewslettersForUser :exec
DELETE FROM newsletter WHERE user_id = $1;

-- name: UpdateNewsletterStatus :exec
UPDATE newsletter SET status = $1 WHERE id = $2 AND user_id = $3;

-- name: GetNewsletterByUnsubscribeToken :one
SELECT nl.id AS newsletter_id, nl.name, u.email, u.id AS user_id FROM newsletter AS nl
INNER JOIN app_user AS u ON nl.user_id = u.id
WHERE unsubscribe_token = $1;

-- name: DeactivateNewsletterByUnsubscribeToken :exec
UPDATE newsletter SET status = 'inactive', unsubscribe_token = gen_random_uuid() WHERE unsubscribe_token = $1;