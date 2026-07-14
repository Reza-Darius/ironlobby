package database

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// domain errors to not expose PG internals
var (
	ErrUserNotFound = errors.New("user not found")
	ErrLobbyExists  = errors.New("lobby already exists")
	ErrUserExists   = errors.New("user already exists")
)

func (db Database) OpenLobbies(ctx context.Context) (int64, error) {
	q := New(db.pool)
	return q.OpenLobbies(ctx)
}

func (db Database) GetUser(ctx context.Context, id uuid.UUID) (string, error) {
	q := New(db.pool)
	return q.GetPlayer(ctx, id)
}

func (db Database) NewUser(ctx context.Context, playerName string) (uuid.UUID, error) {
	q := New(db.pool)
	player, err := q.InsertNewPlayer(ctx, playerName)
	if err != nil {
		pgErr, e := errors.AsType[*pgconn.PgError](err)
		if e {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				{
					err = ErrUserExists
				}
			}
		}
	}
	return player.ID, err
}

func (db Database) CreateLobby(ctx context.Context, arg InsertLobbyParams) (Lobby, error) {
	q := New(db.pool)
	lobby, err := q.InsertLobby(ctx, arg)
	if err != nil {
		pgErr, e := errors.AsType[*pgconn.PgError](err)
		if e {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				{
					err = ErrLobbyExists
				}
			}
		}
	}
	return lobby, err
}

func (db Database) JoinLobby(ctx context.Context, arg AssignPlayerToLobbyParams) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	q := New(db.pool)
	qtx := q.WithTx(tx)

	err = qtx.AssignPlayerToLobby(ctx, arg)
	if err != nil {
		return err
	}

	err = qtx.IncrementCountry(ctx, IncrementCountryParams{
		LobbyID:    arg.LobbyID,
		CountryTag: arg.CountryTag,
	})
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

type LobbyInfo struct {
	Lobby     Lobby                  `json:"lobby"`
	Countries []GetLobbyCountriesRow `json:"countries"`
	Players   []GetLobbyPlayersRow   `json:"players"`
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
		Lobby:     lobby,
		Countries: lobbyCountries,
		Players:   lobbyPlayer,
	}, nil
}
