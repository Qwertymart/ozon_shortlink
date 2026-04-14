-- +goose Up
CREATE TABLE IF NOT EXISTS urls (
    id BIGSERIAL PRIMARY KEY,
    short_url VARCHAR(10) UNIQUE NOT NULL,
    original_url TEXT UNIQUE NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS urls;
