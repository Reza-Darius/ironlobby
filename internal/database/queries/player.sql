-- name: GetAllPlayer :many
SELECT * FROM player;

-- name: InsertPlayer :one
INSERT INTO player(player_name) 
VALUES ($1)
RETURNING *;

