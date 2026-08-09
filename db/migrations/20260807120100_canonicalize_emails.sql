-- migrate:up
UPDATE app_user SET email = lower(email), pending_email = lower(pending_email);
UPDATE white_listed_email SET email = lower(email);

-- Belt and braces: the application lowercases on write, and these indexes make
-- a bug there a constraint violation rather than a duplicate account.
CREATE UNIQUE INDEX app_user_email_lower_key ON app_user (lower(email));
CREATE UNIQUE INDEX white_listed_email_lower_key ON white_listed_email (lower(email));

-- migrate:down
DROP INDEX white_listed_email_lower_key;
DROP INDEX app_user_email_lower_key;
