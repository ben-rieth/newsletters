-- migrate:up

-- auth
CREATE TABLE app_user (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE refresh_token (
    id BIGSERIAL PRIMARY KEY,
    token TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_id UUID NOT NULL REFERENCES app_user(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ
);

-- business

CREATE TABLE feed (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    url TEXT UNIQUE NOT NULL,
    last_retrieved_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TYPE feed_url_source AS ENUM ('canonical', 'user_submitted', 'in_feed_response', 'unknown');

CREATE TABLE feed_url (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    feed_id UUID NOT NULL REFERENCES feed(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    url TEXT UNIQUE NOT NULL,
    source feed_url_source NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE feed_item (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    feed_id UUID NOT NULL REFERENCES feed(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    title TEXT NOT NULL,
    url TEXT NOT NULL,
    publish_date TIMESTAMPTZ NOT NULL,
    retrieved_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


CREATE TYPE frequency AS ENUM ('monthly', 'weekly', 'daily');

CREATE TABLE newsletter (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    frequency frequency NOT NULL,
    send_day INT NOT NULL DEFAULT 0,
    send_hour INT NOT NULL,
    send_minute INT NOT NULL DEFAULT 0,
    send_timezone TEXT NOT NULL DEFAULT 'UTC',
    last_sent_at TIMESTAMPTZ,
    next_send_time TIMESTAMPTZ NOT NULL,
    user_id UUID NOT NULL REFERENCES app_user(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE newsletter_feed (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    newsletter_id UUID NOT NULL REFERENCES newsletter(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    feed_id UUID NOT NULL REFERENCES feed(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    user_id UUID NOT NULL REFERENCES app_user(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    alias TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TYPE item_state AS ENUM ('read', 'unread');

CREATE TABLE newsletter_feed_item_status (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    newsletter_feed_id UUID NOT NULL REFERENCES newsletter_feed(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    item_id UUID NOT NULL REFERENCES feed_item(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    user_id UUID NOT NULL REFERENCES app_user(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    state item_state NOT NULL DEFAULT 'unread',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TYPE filter_field AS ENUM ('title', 'url');
CREATE TYPE filter_operator As ENUM ('contains', 'does_not_contain');

CREATE TABLE newsletter_feed_filter (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    newsletter_feed_id UUID NOT NULL REFERENCES newsletter_feed(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    user_id UUID NOT NULL REFERENCES app_user(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    field filter_field NOT NULL,
    operator filter_operator NOT NULL,
    pattern TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- migrate:down

DROP TABLE refresh_token;
DROP TABLE newsletter_feed_filter;
DROP TYPE filter_operator;
DROP TYPE filter_field;
DROP TABLE newsletter_feed_item_status;
DROP TABLE newsletter_feed;
DROP TABLE newsletter;
DROP TABLE feed_item;
DROP TABLE feed_url;
DROP TABLE feed;
DROP TABLE app_user;
DROP TYPE frequency;
DROP TYPE item_state;
DROP TYPE feed_url_source;
