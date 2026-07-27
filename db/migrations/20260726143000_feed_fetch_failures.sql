-- migrate:up
CREATE TYPE feed_fetch_failure_kind AS ENUM ('http_status', 'transport', 'parse', 'unsafe_url');

CREATE TABLE feed_fetch_failure (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    feed_id UUID REFERENCES feed(id) ON UPDATE RESTRICT ON DELETE CASCADE,
    url TEXT NOT NULL,
    kind feed_fetch_failure_kind NOT NULL,
    status_code INT,
    message TEXT NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX feed_fetch_failure_feed_occurred_idx ON feed_fetch_failure (feed_id, occurred_at DESC);
CREATE INDEX feed_fetch_failure_url_occurred_idx ON feed_fetch_failure (url, occurred_at DESC);

ALTER TABLE feed ADD COLUMN disabled_until TIMESTAMPTZ;
ALTER TABLE feed ADD COLUMN disable_count INT NOT NULL DEFAULT 0;

-- migrate:down
ALTER TABLE feed DROP COLUMN disable_count;
ALTER TABLE feed DROP COLUMN disabled_until;

DROP TABLE feed_fetch_failure;
DROP TYPE feed_fetch_failure_kind;
