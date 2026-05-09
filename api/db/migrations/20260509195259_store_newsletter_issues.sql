-- migrate:up
CREATE TABLE newsletter_issue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    newsletter_id UUID NOT NULL REFERENCES newsletter(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    user_id UUID NOT NULL REFERENCES app_user(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE newsletter_feed_item_status ADD COLUMN issue_id UUID NOT NULL REFERENCES newsletter_issue(id) ON UPDATE RESTRICT ON DELETE RESTRICT;
ALTER TABLE newsletter_feed_item_status DROP COLUMN newsletter_feed_id;

-- migrate:down
ALTER TABLE newsletter_feed_item_status ADD COLUMN newsletter_feed_id UUID NOT NULL REFERENCES newsletter_feed(id) ON UPDATE RESTRICT ON DELETE RESTRICT;
ALTER TABLE newsletter_feed_item_status DROP COLUMN issue_id;
DROP TABLE newsletter_issue;
