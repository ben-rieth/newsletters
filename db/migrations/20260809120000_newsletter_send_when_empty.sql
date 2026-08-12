-- migrate:up
ALTER TABLE newsletter ADD COLUMN send_when_empty BOOLEAN NOT NULL DEFAULT FALSE;

-- migrate:down
ALTER TABLE newsletter DROP COLUMN send_when_empty;
