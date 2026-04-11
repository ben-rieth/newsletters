-- migrate:up

CREATE TYPE Frequency AS ENUM ('monthly', 'weekly', 'daily');

CREATE TABLE newsletter (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    frequency Frequency NOT NULL,
    send_day INT NOT NULL DEFAULT 0,
    send_hour INT NOT NULL,
    send_minute INT NOT NULL DEFAULT 0,
    send_timezone TEXT NOT NULL DEFAULT 'UTC',
    last_sent_at TIMESTAMPTZ,
    next_send_time TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE feed (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    newsletter_id UUID NOT NULL REFERENCES newsletter(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    last_retrieved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- migrate:down

DROP TABLE feed;
DROP TABLE newsletter;
DROP TYPE Frequency;
