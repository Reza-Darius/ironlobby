-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS player (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_name TEXT NOT NULL UNIQUE
);

CREATE TYPE GAMEMODE AS ENUM ('vanilla', 'modded', 'rp');

CREATE TABLE IF NOT EXISTS lobby (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    host_player UUID NOT NULL REFERENCES player (id) ON DELETE CASCADE,
    lobby_name TEXT NOT NULL,
    start_at TIMESTAMPTZ NOT NULL,
    player_count INT NOT NULL,
    gamemode GAMEMODE NOT NULL,
    ingame_id TEXT,

    CHECK (player_count BETWEEN 1 AND 32)
);

CREATE TABLE IF NOT EXISTS player_lobby (
    player_id UUID REFERENCES player (id) ON DELETE CASCADE,
    lobby_id UUID NOT NULL REFERENCES lobby (id) ON DELETE CASCADE,
    country_id INTEGER NOT NULL REFERENCES countries (id),
    joined_at TIMESTAMPTZ NOT NULL,

    PRIMARY KEY (player_id)
);

CREATE TABLE IF NOT EXISTS lobby_countries (
    lobby_id UUID REFERENCES lobby (id) ON DELETE CASCADE,
    country_id INT REFERENCES countries (id) ON DELETE SET NULL,
    max_slots INT NOT NULL DEFAULT 1,

    CHECK (max_slots BETWEEN 1 AND 32),
    PRIMARY KEY (lobby_id, country_id)
);

CREATE TABLE IF NOT EXISTS countries (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    country_tag CHAR(3) NOT NULL UNIQUE
);

INSERT INTO countries (country_tag) VALUES
    ('GER'), -- Germany
    ('SOV'), -- Soviet Union
    ('USA'), -- United States
    ('ENG'), -- United Kingdom
    ('FRA'), -- France
    ('ITA'), -- Italy
    ('JAP'), -- Japan
    ('CHI'), -- China
    ('POL'), -- Poland
    ('CAN'), -- Canada
    ('AST'), -- Australia
    ('NZL'), -- New Zealand
    ('SAF'), -- South Africa
    ('RAJ'), -- British Raj
    ('HOL'), -- Netherlands
    ('BEL'), -- Belgium
    ('LUX'), -- Luxembourg
    ('NOR'), -- Norway
    ('DEN'), -- Denmark
    ('SWE'), -- Sweden
    ('FIN'), -- Finland
    ('SPR'), -- Spain
    ('POR'), -- Portugal
    ('TUR'), -- Turkey
    ('GRE'), -- Greece
    ('YUG'), -- Yugoslavia
    ('ROM'), -- Romania
    ('HUN'), -- Hungary
    ('BUL'), -- Bulgaria
    ('CZE'), -- Czechoslovakia
    ('MEX'), -- Mexico
    ('BRA'), -- Brazil
    ('ARG'), -- Argentina
    ('CHL'), -- Chile
    ('PRC'), -- Communist China
    ('MAN'), -- Manchukuo
    ('MON'), -- Mongolia
    ('TIB'), -- Tibet
    ('PER'), -- Iran/Persia
    ('IRQ'), -- Iraq
    ('EGY'), -- Egypt
    ('ETH'), -- Ethiopia
    ('THA'), -- Thailand
    ('PHI'); -- Philippines

-- +goose Down
SELECT 'down SQL query';
