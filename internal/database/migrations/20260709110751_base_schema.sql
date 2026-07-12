-- +goose Up
CREATE TABLE IF NOT EXISTS test (
  id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name TEXT
);

-- +goose Down
SELECT 'down SQL query';
