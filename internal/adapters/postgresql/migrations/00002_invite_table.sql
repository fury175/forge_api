-- +goose Up
CREATE TABLE IF NOT EXISTS invites(
    id          BIGSERIAL   PRIMARY KEY,
    user_id     BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invite_id   TEXT        NOT NULL,
    expiry        TIMESTAMPTZ NOT NULL DEFAULT (now() + INTERVAL '3 hours'),
    created_at  TIMESTAMPTZ DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS invites;
