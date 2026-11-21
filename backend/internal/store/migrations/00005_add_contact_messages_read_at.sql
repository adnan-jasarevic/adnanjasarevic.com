-- +goose Up
ALTER TABLE contact_messages ADD COLUMN read_at timestamptz;

-- +goose Down
ALTER TABLE contact_messages DROP COLUMN read_at;
