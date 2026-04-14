-- name: CreateRefreshToken :exec
INSERT INTO refresh_token (token, user_id, expires_at) VALUES ($1, $2, $3);

-- name: GetRefreshToken :one
SELECT token, revoked_at, expires_at, user_id FROM refresh_token WHERE token = $1;

-- name: RevokeToken :exec
UPDATE refresh_token SET revoked_at = NOW() WHERE token = $1;

-- name: DeleteAllRefreshTokensForUser :exec
DELETE FROM refresh_token WHERE user_id = $1;
