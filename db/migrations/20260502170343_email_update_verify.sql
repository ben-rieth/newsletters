-- migrate:up
ALTER TABLE app_user ADD COLUMN pending_email TEXT NOT NULL DEFAULT '';

-- migrate:down
ALTER TABLE app_user REMOVE COLUMN pending_email;

