-- migrate:up
ALTER TABLE feed ADD COLUMN html_url TEXT NOT NULL DEFAULT '';

-- migrate:down
ALTER TABLE feed REMOVE COLUMN html_url;
