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

-- name: FindValidToken :one
SELECT * FROM verification_token 
WHERE user_id = $1 AND purpose = $2 AND code = $3 
AND expires_at > @expires_at_greater_than;