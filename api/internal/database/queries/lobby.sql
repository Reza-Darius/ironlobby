-- name: GetLobbyInfo :one
SELECT
  *
FROM lobby WHERE id = $1;

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
    lobby_countries.occupied_slots,
    lobby_countries.max_slots
FROM lobby_countries
JOIN countries ON lobby_countries.country_id = countries.id
WHERE lobby_countries.lobby_id = $1;

-- name: InsertLobby :one
INSERT INTO lobby(
    host_player,
    lobby_name,
    starts_at,
    gamemode
) 
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateGameId :exec
UPDATE lobby SET ingame_id = $1 WHERE id = $2;

-- name: IncrementPlayerCount :exec
UPDATE lobby SET player_count = player_count + 1 WHERE id = $1;

-- name: DecrementPlayerCount :exec
UPDATE lobby SET player_count = player_count - 1 WHERE id = $1;

-- name: AddLobbyCountry :exec
INSERT INTO lobby_countries (lobby_id, country_id, max_slots)
VALUES (
    $1,
    (SELECT id FROM countries WHERE country_tag = $2),
    $3
);

-- name: OpenLobbies :one
SELECT COUNT(*) FROM lobby WHERE starts_at <= NOW();
