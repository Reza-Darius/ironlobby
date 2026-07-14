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
	ErrUserNotFound          = errors.New("user not found")
	ErrLobbyExists           = errors.New("lobby already exists")
	ErrUserExists            = errors.New("user already exists")
	ErrNationSlotsFull       = errors.New("the requested nation's slots are full")
	ErrPlayerAlreadyAssigned = errors.New("player is already assigned to Nation")
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

// this function adds a player to a lobby or changes the player's country tag inside the lobby
func (db Database) JoinLobby(ctx context.Context, arg UpsertPlayerLobbyParams) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := New(db.pool).WithTx(tx)

	targetTag := arg.CountryTag

	// we lock the targeted country tag
	maxSlots, err := q.LockLobbyCountrySlot(ctx, LockLobbyCountrySlotParams{
		LobbyID:    arg.LobbyID,
		CountryTag: targetTag,
	})
	if err != nil {
		return err
	}

	// check limit by counting all the listed entried for the lobby
	// notably: this excludes the requesting player
	numOccuppied, err := q.CountCountryOccupants(ctx, CountCountryOccupantsParams{
		PlayerID:   arg.PlayerID,
		LobbyID:    arg.LobbyID,
		CountryTag: targetTag,
	})
	if err != nil {
		return err
	}

	// if the player is already on that tag inside the lobby this always passes
	if numOccuppied >= int64(maxSlots) {
		return ErrNationSlotsFull
	}

	_, err = q.UpsertPlayerLobby(ctx, UpsertPlayerLobbyParams{
		CountryTag: targetTag,
		PlayerID:   arg.PlayerID,
		LobbyID:    arg.LobbyID,
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
