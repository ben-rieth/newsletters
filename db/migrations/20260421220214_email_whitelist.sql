-- migrate:up
CREATE TABLE white_listed_email (
    email TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- migrate:down
DROP TABLE white_listed_email;