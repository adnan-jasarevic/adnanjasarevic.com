-- +goose Up
CREATE TABLE projects (
    id          bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title       text        NOT NULL,
    slug        text        NOT NULL UNIQUE,
    summary     text        NOT NULL DEFAULT '',
    description text        NOT NULL DEFAULT '',
    tech_stack  text[]      NOT NULL DEFAULT '{}',
    repo_url    text,
    live_url    text,
    sort_order  integer     NOT NULL DEFAULT 0,
    published   boolean     NOT NULL DEFAULT false,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE projects;
