-- +goose Up
ALTER TABLE metrics ADD COLUMN IF NOT EXISTS hash VARCHAR(64);

-- +goose Down
ALTER TABLE metrics DROP COLUMN IF EXISTS hash;
