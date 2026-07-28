-- migrate:up
CREATE INDEX issue_item_issue_id_state_idx ON issue_item (issue_id, state);

-- migrate:down
DROP INDEX issue_item_issue_id_state_idx;
