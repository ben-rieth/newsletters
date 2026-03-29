-- migrate:up

CREATE TYPE Frequency AS ENUM ('monthly', 'weekly', 'daily');

CREATE TABLE newsletters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    frequency Frequency NOT NULL,
    send_day INT NOT NULL DEFAULT 0,
    send_hour INT NOT NULL,
    send_minute INT NOT NULL DEFAULT 0,
    last_sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- migrate:down

DROP TABLE newsletters;
DROP TYPE Frequency;