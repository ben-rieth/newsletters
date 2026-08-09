-- name: CreateRefreshToken :exec
INSERT INTO refresh_token (token, user_id, expires_at) VALUES ($1, $2, $3);

-- name: GetRefreshToken :one
SELECT token, revoked_at, expires_at, user_id FROM refresh_token WHERE token = $1;

-- name: RevokeToken :exec
UPDATE refresh_token SET revoked_at = NOW(), updated_at = NOW() WHERE token = $1;

-- name: DeleteAllRefreshTokensForUser :exec
DELETE FROM refresh_token WHERE user_id = $1;

-- name: SaveVerificationToken :exec
INSERT INTO verification_token (user_id, code, purpose, expires_at)
VALUES ($1, $2, $3, $4);

-- name: DeleteExistingTokensWithPurpose :exec
DELETE FROM verification_token WHERE user_id = $1 AND purpose = $2;

-- name: DeleteAllVerificationTokensForUser :exec
DELETE FROM verification_token WHERE user_id = $1;

-- name: FindUnexpiredToken :one
SELECT * FROM verification_token
WHERE user_id = $1 AND purpose = $2
AND expires_at > @expires_at_greater_than;

-- name: RecordFailedTokenAttempt :one
UPDATE verification_token SET attempts = attempts + 1 WHERE id = $1 RETURNING attempts;

-- name: DeleteVerificationToken :exec
DELETE FROM verification_token WHERE id = $1;