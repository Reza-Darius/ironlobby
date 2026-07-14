-- name: GetAllPlayer :many
SELECT * FROM player;

-- name: GetPlayer :one
SELECT player_name FROM player
WHERE id = $1;

-- name: InsertNewPlayer :one
INSERT INTO player (player_name)
VALUES ($1)
RETURNING *;
