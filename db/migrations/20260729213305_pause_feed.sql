-- migrate:up
ALTER TABLE newsletter_feed ADD COLUMN status newsletter_status NOT NULL DEFAULT 'active';

-- migrate:down
ALTER TABLE newsletter_feed DROP COLUMN status;
