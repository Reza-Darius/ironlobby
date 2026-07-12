-- name: GetAllPlayer :many
SELECT * FROM Player;

-- name: InsertPlayer :exec
INSERT INTO Player(name) 
VALUES ($1);
