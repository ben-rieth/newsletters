-- migrate:up
CREATE TYPE feed_import_state AS ENUM ('pending', 'failed');

CREATE TABLE newsletter_feed_import (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    newsletter_id UUID NOT NULL REFERENCES newsletter(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    user_id UUID NOT NULL REFERENCES app_user(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    url TEXT NOT NULL,
    alias TEXT NOT NULL DEFAULT '',
    status newsletter_status NOT NULL DEFAULT 'active',
    filters JSONB NOT NULL DEFAULT '[]',
    state feed_import_state NOT NULL DEFAULT 'pending',
    error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- migrate:down
DROP TABLE newsletter_feed_import;
DROP TYPE feed_import_state;
