-- name: CreateUser :one
INSERT INTO app_user (email, password) VALUES ($1, $2) RETURNING id;

-- name: DoesUserWithEmailExist :one
SELECT EXISTS(SELECT 1 FROM app_user WHERE email = $1);

-- name: GetUserByEmail :one
SELECT * FROM app_user WHERE email = $1;

-- name: GetUserById :one
SELECT * FROM app_user WHERE id = $1;

-- name: UpdateUserEmail :exec
UPDATE app_user SET email = $1, updated_at = NOW() WHERE id = $2;

-- name: UpdateUserPassword :exec
UPDATE app_user SET password = $1, updated_at = NOW() WHERE id = $2;

-- name: DeleteUser :exec
DELETE FROM app_user WHERE id = $1;

-- name: MarkUserEmailAsVerified :exec
UPDATE app_user SET email_verified_at = NOW() WHERE id = $1;

-- name: MarkUserEmailUpdateAsVerified :exec
UPDATE app_user SET email_verified_at = NOW(), email = pending_email, pending_email = ''
WHERE id = $1;

-- name: IsWhiteListedEmail :one
SELECT EXISTS(SELECT 1 FROM white_listed_email WHERE email = $1);

-- name: AddPendingEmailUpdate :exec
UPDATE app_user SET pending_email = $1, updated_at = NOW() WHERE id = $2;