package database

import (
	"context"

	"github.com/google/uuid"
)

func (db Database) OpenLobbies(ctx context.Context) (int64, error) {
	q := New(db.pool)
	return q.OpenLobbies(ctx)
}

func (db Database) NewUser(ctx context.Context, playerName string) (uuid.UUID, error) {
	q := New(db.pool)
	player, err := q.InsertNewPlayer(ctx, playerName)
	return player.ID, err
}

type LobbyInfo struct {
	Lobby Lobby `json:"lobby"`
	Countries []GetLobbyCountriesRow `json:"countries"`
	Players []GetLobbyPlayersRow `json:"players"`
}

func (db Database) GetLobby(ctx context.Context, lobbyID int64) (LobbyInfo, error) {
	q := New(db.pool)
	lobby, err := q.GetLobbyInfo(ctx, lobbyID)
	if err != nil {
		return LobbyInfo{}, err
	}
	lobbyCountries, err := q.GetLobbyCountries(ctx, lobbyID)
	if err != nil {
		return LobbyInfo{}, err
	}
	lobbyPlayer, err := q.GetLobbyPlayers(ctx, lobbyID)
	if err != nil {
		return LobbyInfo{}, err
	}
	return LobbyInfo{
		Lobby: lobby,
		Countries: lobbyCountries,
		Players: lobbyPlayer,
	}, nil
}
