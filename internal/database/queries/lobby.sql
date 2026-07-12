-- name: GetLobby :one
SELECT * FROM lobby WHERE id = $1;

-- name: InsertLobby :one
INSERT INTO lobby(host_player, lobby_name, start_at, player_count, gamemode) 
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateGameId :exec
UPDATE lobby SET ingame_id = $1 WHERE id = $2;
