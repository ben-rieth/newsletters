-- migrate:up

-- Access tokens are stateless, so revoking one means refusing to honour tokens
-- issued before a cutoff rather than deleting a row. Defaulting to NOW() signs
-- every existing session out once.
ALTER TABLE app_user
    ADD COLUMN sessions_valid_from TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- These budgets live on the account because the alternatives leak: a per-token
-- budget resets whenever the user asks for a new code, and a per-IP one is free
-- to dodge for anyone holding an IPv6 allocation.
ALTER TABLE app_user
    ADD COLUMN verify_attempts INT NOT NULL DEFAULT 0,
    ADD COLUMN verify_locked_until TIMESTAMPTZ,
    ADD COLUMN failed_signin_attempts INT NOT NULL DEFAULT 0,
    ADD COLUMN signin_locked_until TIMESTAMPTZ;

-- migrate:down

ALTER TABLE app_user
    DROP COLUMN sessions_valid_from,
    DROP COLUMN verify_attempts,
    DROP COLUMN verify_locked_until,
    DROP COLUMN failed_signin_attempts,
    DROP COLUMN signin_locked_until;
