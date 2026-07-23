-- migrate:up
ALTER TABLE newsletter ADD COLUMN original_next_send_time TIMESTAMPTZ;

-- migrate:down
ALTER TABLE newsletter DROP COLUMN original_next_send_time;

