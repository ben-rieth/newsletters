-- migrate:up
-- The scheduler already sent these on the 1st and modulo 7 respectively; this
-- makes the stored day match so the API can stop accepting the old values.
UPDATE newsletter SET send_day = 1 WHERE frequency = 'monthly' AND send_day < 1;
UPDATE newsletter SET send_day = send_day % 7 WHERE frequency = 'weekly' AND send_day > 6;

-- migrate:down
