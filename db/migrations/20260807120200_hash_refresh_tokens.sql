-- migrate:up
-- refresh_token.token now holds a SHA-256 hash of the secret handed to the
-- client, so a database leak no longer yields usable sessions. Existing rows
-- store raw secrets that can't be converted without the plaintext; dropping
-- them just forces everyone to sign in again.
DELETE FROM refresh_token;

-- migrate:down
DELETE FROM refresh_token;
