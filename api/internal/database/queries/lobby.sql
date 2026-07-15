-- name: GetLobbyInfo :one
SELECT *
FROM lobby
WHERE id = $1;

-- name: GetLobbyPlayers :many
SELECT
    player.player_name,
    player_lobby.lobby_id,
    countries.country_tag
FROM player_lobby
JOIN countries ON player_lobby.country_id = countries.id
JOIN player ON player_lobby.player_id = player.id
WHERE player_lobby.lobby_id = $1;

-- name: GetLobbyCountries :many
SELECT
    countries.country_tag,
    lobby_countries.max_slots
FROM lobby_countries
JOIN countries ON lobby_countries.country_id = countries.id
WHERE lobby_countries.lobby_id = $1;

-- name: InsertLobby :one
INSERT INTO lobby (
    host_player,
    lobby_name,
    starts_at,
    gamemode,
    description
)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateGameId :exec
UPDATE lobby SET ingame_id = $1
WHERE id = $2;

-- name: IncrementPlayerCount :one
UPDATE lobby
SET player_count = player_count + 1
WHERE
    id = $1
    AND player_count <= 32
RETURNING player_count;

-- name: DecrementPlayerCount :one
UPDATE lobby SET player_count = player_count - 1
WHERE id = $1 RETURNING player_count;

-- name: UpsertLobbyCountry :one
INSERT INTO lobby_countries (lobby_id, country_id, max_slots)
SELECT
    $1,
    id,
    $3
FROM countries
WHERE country_tag = $2
ON CONFLICT (lobby_id, country_id)
DO UPDATE SET max_slots = excluded.max_slots
RETURNING *;


-- name: OpenLobbies :one
SELECT COUNT(*) FROM lobby
WHERE starts_at <= NOW();

-- name: GetLobbyFromHostID :one
SELECT l.id FROM lobby AS l
WHERE l.host_player = $1 AND l.id = $2;

-- name: SearchPlayerInLobby :one
SELECT country_id FROM player_lobby
WHERE lobby_id = $1 AND player_id = $2;

-- name: LockLobbyCountrySlot :one
SELECT lc.max_slots
FROM lobby_countries AS lc
WHERE
    lobby_id = $1 AND country_id = (
        SELECT c.id FROM countries AS c
        WHERE c.country_tag = $2
    )
FOR UPDATE OF lc;

-- name: CountCountryOccupants :one
SELECT COUNT(*) FROM player_lobby
WHERE
    lobby_id = $1 AND country_id
    = (
        SELECT id FROM countries
        WHERE country_tag = $2
    )
    AND player_id != $3;

-- name: UpsertPlayerLobby :one
INSERT INTO player_lobby (player_id, lobby_id, country_id)
SELECT
    $1,
    $2,
    id
FROM countries
WHERE country_tag = $3
ON CONFLICT (lobby_id, player_id)
DO UPDATE SET country_id = excluded.country_id
RETURNING *;

-- name: UnassignPlayer :exec
DELETE FROM player_lobby
WHERE lobby_id = $1 AND player_id = $2;
