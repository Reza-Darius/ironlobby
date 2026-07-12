-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS player (
  id UUDI PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL UNIQUE
);

CREATE TYPE gamemode AS ENUM ('vanilla', 'modded', 'rp');

CREATE TABLE IF NOT EXISTS lobby (
  id UUDI PRIMARY KEY DEFAULT gen_random_uuid(),
  host_player UUID NOT NULL REFERENCES player(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  start_at TIMESTAMPTZ NOT NULL,
  players INT,
  mode gamemode,
  ingame_id TEXT,

  CHECK (players BETWEEN 1 AND 32)
);

CREATE TABLE IF NOT EXISTS player_country (
  player_id UUID NOT NULL PRIMARY KEY REFERENCES player(id) ON DELETE CASCADE,
  lobby_id UUDI NOT NULL REFERENCES lobby(id) ON DELETE CASCADE,
  country_id INTEGER NOT NULL REFERENCES countries(id),
);

CREATE TABLE IF NOT EXISTS lobby_countries (
  lobby_id UUID NOT NULL REFERENCES lobby(id) ON DELETE CASCADE,
  country_id INT NOT NULL UNIQUE REFERENCES countries(id) ON DELETE SET NULL,
  max_slots INT NOT NULL DEFAULT 1,

  CHECK (max_slots BETWEEN 1 AND 32)
  PRIMARY KEY (lobby_id, country_id),
);

CREATE TABLE IF NOT EXISTS countries (
  id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  name TEXT NOT NULL,
);
- +goose Down
SELECT 'down SQL query';
