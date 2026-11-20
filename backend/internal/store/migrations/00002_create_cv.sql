-- +goose Up
CREATE TABLE profile (
    id           boolean     PRIMARY KEY DEFAULT true CHECK (id),
    full_name    text        NOT NULL,
    headline     text        NOT NULL DEFAULT '',
    summary      text        NOT NULL DEFAULT '',
    location     text        NOT NULL DEFAULT '',
    email        text,
    github_url   text,
    linkedin_url text,
    updated_at   timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE experience (
    id          bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    company     text        NOT NULL,
    role        text        NOT NULL,
    location    text        NOT NULL DEFAULT '',
    start_date  date        NOT NULL,
    end_date    date        CHECK (end_date IS NULL OR end_date >= start_date),
    description text        NOT NULL DEFAULT '',
    highlights  text[]      NOT NULL DEFAULT '{}',
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE education (
    id          bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    institution text        NOT NULL,
    degree      text        NOT NULL,
    field       text        NOT NULL DEFAULT '',
    start_date  date        NOT NULL,
    end_date    date        CHECK (end_date IS NULL OR end_date >= start_date),
    description text        NOT NULL DEFAULT '',
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE skills (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name       text        NOT NULL UNIQUE,
    category   text        NOT NULL DEFAULT '',
    sort_order integer     NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE skills;
DROP TABLE education;
DROP TABLE experience;
DROP TABLE profile;
