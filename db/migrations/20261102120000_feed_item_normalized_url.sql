-- migrate:up

ALTER TABLE feed_item ADD COLUMN normalized_url TEXT;

-- Approximates the Go normalizer (https scheme, no trailing path slash) so most
-- stored items still block a re-fetched copy. Later copies of existing
-- duplicates stay NULL because issues still reference them.
UPDATE feed_item AS fi
SET normalized_url = regexp_replace(
    regexp_replace(fi.url, '^http://', 'https://'),
    '^([^?#]*)/([?#].*)?$', '\1\2'
)
FROM (
    SELECT id, ROW_NUMBER() OVER (
        PARTITION BY feed_id, regexp_replace(
            regexp_replace(url, '^http://', 'https://'),
            '^([^?#]*)/([?#].*)?$', '\1\2'
        )
        ORDER BY created_at, id
    ) AS copy
    FROM feed_item
) AS ranked
WHERE ranked.id = fi.id AND ranked.copy = 1;

CREATE UNIQUE INDEX feed_item_feed_id_normalized_url_key ON feed_item (feed_id, normalized_url);

-- migrate:down

DROP INDEX feed_item_feed_id_normalized_url_key;
ALTER TABLE feed_item DROP COLUMN normalized_url;
