-- migrate:up
ALTER TYPE token_purpose ADD VALUE IF NOT EXISTS 'email_update';

ALTER TABLE verification_token ADD COLUMN attempts INT NOT NULL DEFAULT 0;

-- migrate:down
ALTER TABLE verification_token DROP COLUMN attempts;

DELETE FROM verification_token WHERE purpose = 'email_update';

ALTER TYPE token_purpose RENAME TO token_purpose_old;

CREATE TYPE token_purpose AS ENUM ('email_verify');

ALTER TABLE verification_token
    ALTER COLUMN purpose TYPE token_purpose USING purpose::text::token_purpose;

DROP TYPE token_purpose_old;
