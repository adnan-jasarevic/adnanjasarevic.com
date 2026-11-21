-- +goose Up
CREATE TABLE admin_sessions (
    token_hash   bytea       PRIMARY KEY,
    github_id    bigint      NOT NULL,
    github_login text        NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    expires_at   timestamptz NOT NULL
);

CREATE INDEX admin_sessions_expires_at_idx ON admin_sessions (expires_at);

-- +goose Down
DROP TABLE admin_sessions;
