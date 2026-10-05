-- migrate:up
ALTER TABLE app_user ADD COLUMN issue_retention_days INTEGER NOT NULL DEFAULT 0;

-- migrate:down
ALTER TABLE app_user DROP COLUMN issue_retention_days;
