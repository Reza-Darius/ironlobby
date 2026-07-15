-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS player (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS countries (
    id SMALLINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    country_tag CHAR(3) NOT NULL UNIQUE,
    country_name TEXT NOT NULL UNIQUE
);

INSERT INTO countries (country_tag, country_name) VALUES
('GER', 'Germany'),
('SOV', 'Soviet Union'),
('USA', 'United States'),
('ENG', 'United Kingdom'),
('FRA', 'France'),
('ITA', 'Italy'),
('JAP', 'Japan'),
('CHI', 'China'),
('POL', 'Poland'),
('CAN', 'Canada'),
('AST', 'Australia'),
('NZL', 'New Zealand'),
('SAF', 'South Africa'),
('RAJ', 'British Raj'),
('HOL', 'Netherlands'),
('BEL', 'Belgium'),
('LUX', 'Luxembourg'),
('NOR', 'Norway'),
('DEN', 'Denmark'),
('SWE', 'Sweden'),
('FIN', 'Finland'),
('SPR', 'Spain'),
('POR', 'Portugal'),
('TUR', 'Turkey'),
('GRE', 'Greece'),
('YUG', 'Yugoslavia'),
('ROM', 'Romania'),
('HUN', 'Hungary'),
('BUL', 'Bulgaria'),
('CZE', 'Czechoslovakia'),
('MEX', 'Mexico'),
('BRA', 'Brazil'),
('ARG', 'Argentina'),
('CHL', 'Chile'),
('PRC', 'Communist China'),
('MAN', 'Manchukuo'),
('MON', 'Mongolia'),
('TIB', 'Tibet'),
('PER', 'Iran'),
('IRQ', 'Iraq'),
('EGY', 'Egypt'),
('ETH', 'Ethiopia'),
('THA', 'Thailand'),
('PHI', 'Philippines');

CREATE TYPE GAMEMODE AS ENUM ('vanilla', 'modded', 'rp');

CREATE TABLE IF NOT EXISTS lobby (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    host_player UUID NOT NULL REFERENCES player (id) ON DELETE CASCADE,
    lobby_name TEXT NOT NULL,
    starts_at TIMESTAMPTZ NOT NULL,
    gamemode GAMEMODE NOT NULL,
    ingame_id CHAR(17),
    description VARCHAR(1000) NOT NULL
);

CREATE TABLE IF NOT EXISTS lobby_countries (
    lobby_id BIGINT REFERENCES lobby (id) ON DELETE CASCADE,
    country_id SMALLINT REFERENCES countries (id) ON DELETE CASCADE,
    max_slots SMALLINT NOT NULL DEFAULT 1,

    CHECK (max_slots <= 32),
    PRIMARY KEY (lobby_id, country_id)
);

CREATE TABLE player_lobby (
    player_id UUID NOT NULL REFERENCES player (id) ON DELETE CASCADE,
    lobby_id BIGINT NOT NULL,
    country_id SMALLINT NOT NULL,
    note TEXT,

    joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (lobby_id, player_id),
    FOREIGN KEY (lobby_id, country_id)
    REFERENCES lobby_countries (lobby_id, country_id)
    ON DELETE CASCADE
);

-- +goose Down
SELECT 'down SQL query';
