-- name: GetAllPlayer :many
SELECT * FROM player;

-- name: InsertNewPlayer :one
INSERT INTO player(player_name) 
VALUES ($1)
RETURNING *;

-- name: AssignPlayerToLobby :exec
INSERT INTO player_lobby(player_id, lobby_id, country_id)
VALUES(
  $1,
  $2,
  (SELECT id FROM countries WHERE country_tag = $3)
);

-- name: IncrementCountry :exec
UPDATE lobby_countries SET occupied_slots = occupied_slots + 1 WHERE lobby_id = $1 AND country_id = (
  SELECT id FROM countries WHERE country_tag = $2
);

-- name: UnassignPlayer :exec
DELETE FROM player_lobby WHERE player_id = $1;

-- name: DecrementCountry :exec
UPDATE lobby_countries SET occupied_slots = occupied_slots -1 WHERE lobby_id = $1 AND country_id = (
  SELECT id FROM countries WHERE country_tag = $2
);

