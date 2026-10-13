-- migrate:up

-- The sign-in budget moved into memory and is keyed by account and address
-- together, so an account-wide lock can no longer be armed by a stranger.
ALTER TABLE app_user
    DROP COLUMN failed_signin_attempts,
    DROP COLUMN signin_locked_until;

-- migrate:down

ALTER TABLE app_user
    ADD COLUMN failed_signin_attempts INT NOT NULL DEFAULT 0,
    ADD COLUMN signin_locked_until TIMESTAMPTZ;
