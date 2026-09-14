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

-- The whitelist and uniqueness checks that gated the request go stale while the
-- code sits in the user's inbox, so they are re-asserted where the email actually
-- moves. No row back means the update is no longer valid.
-- name: MarkUserEmailUpdateAsVerified :one
UPDATE app_user AS u SET email_verified_at = NOW(), email = u.pending_email, pending_email = '', updated_at = NOW()
WHERE u.id = $1 AND u.pending_email <> ''
  AND EXISTS (SELECT 1 FROM white_listed_email w WHERE w.email = u.pending_email)
  AND NOT EXISTS (SELECT 1 FROM app_user o WHERE o.email = u.pending_email AND o.id <> u.id)
RETURNING u.id;

-- name: GetUserAuthState :one
SELECT sessions_valid_from FROM app_user WHERE id = $1;

-- name: GetUserVerifyLock :one
SELECT verify_locked_until FROM app_user WHERE id = $1;

-- name: InvalidateUserSessions :exec
UPDATE app_user SET sessions_valid_from = NOW(), updated_at = NOW() WHERE id = $1;

-- Counting up and arming the lock in one statement keeps concurrent wrong
-- guesses from each reading the same pre-increment count.
-- name: RecordFailedVerifyAttempt :one
UPDATE app_user SET
    verify_attempts = CASE
        WHEN verify_attempts + 1 >= sqlc.arg(max_attempts)::int THEN 0
        ELSE verify_attempts + 1
    END,
    verify_locked_until = CASE
        WHEN verify_attempts + 1 >= sqlc.arg(max_attempts)::int THEN sqlc.arg(locked_until)::timestamptz
        ELSE verify_locked_until
    END,
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING verify_attempts, verify_locked_until;

-- name: ResetVerifyAttempts :exec
UPDATE app_user SET verify_attempts = 0, verify_locked_until = NULL, updated_at = NOW()
WHERE id = $1 AND (verify_attempts <> 0 OR verify_locked_until IS NOT NULL);

-- name: CarryVerifyLockout :exec
UPDATE app_user SET verify_attempts = $2, verify_locked_until = $3 WHERE id = $1;

-- name: RecordFailedSignIn :exec
UPDATE app_user SET
    failed_signin_attempts = CASE
        WHEN failed_signin_attempts + 1 >= sqlc.arg(max_attempts)::int THEN 0
        ELSE failed_signin_attempts + 1
    END,
    signin_locked_until = CASE
        WHEN failed_signin_attempts + 1 >= sqlc.arg(max_attempts)::int THEN sqlc.arg(locked_until)::timestamptz
        ELSE signin_locked_until
    END,
    updated_at = NOW()
WHERE id = sqlc.arg(id);

-- name: ResetSignInAttempts :exec
UPDATE app_user SET failed_signin_attempts = 0, signin_locked_until = NULL, updated_at = NOW()
WHERE id = $1 AND (failed_signin_attempts <> 0 OR signin_locked_until IS NOT NULL);

-- name: IsWhiteListedEmail :one
SELECT EXISTS(SELECT 1 FROM white_listed_email WHERE email = $1);

-- name: AddPendingEmailUpdate :exec
UPDATE app_user SET pending_email = $1, updated_at = NOW() WHERE id = $2;