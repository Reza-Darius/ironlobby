-- name: GetAllPlayer :many
SELECT * FROM player;

-- name: GetPlayer :one
SELECT player_name FROM player
WHERE id = $1;

-- name: GetPlayerID :one
SELECT id FROM player
WHERE player_name = $1;

-- name: InsertNewPlayer :one
INSERT INTO player (player_name)
VALUES ($1)
RETURNING *;
