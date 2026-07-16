package database

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// domain errors to not expose PG internals
var (
	ErrUserNotFound          = errors.New("user not found")

	ErrLobbyExists           = errors.New("lobby already exists")
	ErrLobbyDoesntExists     = errors.New("lobby doesnt exist")

	ErrUserExists            = errors.New("user already exists")
	ErrNationSlotsFull       = errors.New("the requested nation's slots are full")
	ErrNationNotAvail        = errors.New("the requested nation is not available")
	ErrPlayerAlreadyAssigned = errors.New("player is already assigned to Nation")
	ErrCountryDoesntExist    = errors.New("provided country tag doesnt exist")
)

func (db *Database) OpenLobbies(ctx context.Context) (int64, error) {
	q := New(db.pool)
	return q.OpenLobbies(ctx)
}

func (db *Database) GetUser(ctx context.Context, id uuid.UUID) (string, error) {
	q := New(db.pool)
	return q.GetPlayer(ctx, id)
}

func (db *Database) NewUser(ctx context.Context, playerName string) (uuid.UUID, error) {
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

type CreateLobbyRequest struct {
	Lobby     InsertLobbyParams          `json:"lobby"`
	Countries []UpsertLobbyCountryParams `json:"countries"`
}

func (db *Database) CreateLobby(ctx context.Context, arg CreateLobbyRequest) (Lobby, error) {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return Lobby{}, err
	}
	defer tx.Rollback(ctx)
	q := New(db.pool).WithTx(tx)

	lobby, err := q.InsertLobby(ctx, arg.Lobby)
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
		return Lobby{}, err
	}

	// batch countries
	batch := &pgx.Batch{}
	for _, country := range arg.Countries {
		batch.Queue(
			`INSERT INTO lobby_countries (lobby_id, country_id, max_slots) VALUES ($1, (SELECT id FROM countries WHERE country_tag = $2), $3)`,
			lobby.ID, country.CountryTag, country.MaxSlots,
		)
	}
	br := tx.SendBatch(ctx, batch)
	defer br.Close()
	for range arg.Countries {
		if _, err := br.Exec(); err != nil {
			return Lobby{}, err
		}
	}
	if err := br.Close(); err != nil {
		return Lobby{}, err
	}
	return lobby, tx.Commit(ctx)
}

// JoinLobby adds a player to a lobby or changes the player's country tag inside the lobby
func (db *Database) JoinLobby(ctx context.Context, arg UpsertLobbyPlayerParams) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := New(db.pool).WithTx(tx)

	// we lock the targeted country tag
	maxSlots, err := q.LockLobbyCountrySlot(ctx, LockLobbyCountrySlotParams{
		LobbyID:    arg.LobbyID,
		CountryTag: arg.CountryTag,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return ErrNationNotAvail
		}
		return err
	}

	// check limit by counting all the listed entried for the lobby
	// notably: this excludes the requesting player
	numOccuppied, err := q.CountCountryOccupants(ctx, CountCountryOccupantsParams{
		PlayerID:   arg.PlayerID,
		LobbyID:    arg.LobbyID,
		CountryTag: arg.CountryTag,
	})
	if err != nil {
		return err
	}

	// if the player is already on that tag inside the lobby this always passes
	if numOccuppied >= int64(maxSlots) {
		return ErrNationSlotsFull
	}

	_, err = q.UpsertLobbyPlayer(ctx, arg)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (db *Database) DeleteLobbyPlayer(ctx context.Context, arg DeleteLobbyPlayerParams) error {
	q := New(db.pool)
	_, err := q.DeleteLobbyPlayer(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ErrUserNotFound
		}
		return err
	}
	return nil
}

type LobbyInfo struct {
	Lobby     Lobby                  `json:"lobby"`
	Countries []GetLobbyCountriesRow `json:"countries"`
	Players   []GetLobbyPlayersRow   `json:"players"`
}

func (db *Database) GetLobby(ctx context.Context, lobbyID int64) (LobbyInfo, error) {
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

func (db *Database) GetLobbyCountries(ctx context.Context, lobbyID int64) ([]GetLobbyCountriesRow, error) {
	q := New(db.pool)
	lobbyCountries, err := q.GetLobbyCountries(ctx, lobbyID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrLobbyDoesntExists
		}
		return nil, err
	}
	return lobbyCountries, nil
}

func (db *Database) GetLobbyPlayers(ctx context.Context, lobbyID int64) ([]GetLobbyPlayersRow, error) {
	q := New(db.pool)
	lobbyPlayer, err := q.GetLobbyPlayers(ctx, lobbyID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrLobbyDoesntExists
		}
		return nil, err
	}
	return lobbyPlayer, nil
}

func (db *Database) PlayerIsHost(ctx context.Context, playerID uuid.UUID, lobbyID int64) (bool, error) {
	q := New(db.pool)
	_, err := q.GetLobbyFromHostID(ctx, GetLobbyFromHostIDParams{
		ID:         lobbyID,
		HostPlayer: playerID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// AddLobbyCountry adds a country to the lobby, or if the tag is alreaddy registered, updates the max slots
func (db *Database) AddLobbyCountry(ctx context.Context, arg UpsertLobbyCountryParams) error {
	q := New(db.pool)
	_, err := q.UpsertLobbyCountry(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ErrCountryDoesntExist
		}
		return err
	}
	return nil
}

func (db *Database) DeleteLobbyCountry(ctx context.Context, arg DeleteLobbyCountryParams) error {
	q := New(db.pool)
	_, err := q.DeleteLobbyCountry(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			// TODO:: better error, this could also mean the lobby doesnt exist
			return ErrCountryDoesntExist
		}
		return err
	}
	return nil
}

func (db *Database) UpdateLobby(ctx context.Context, arg UpdateLobbyParams) error {
	q := New(db.pool)
	_, err := q.UpdateLobby(ctx, arg)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) UpdateLobbyCountry(ctx context.Context, arg UpdateLobbyCountryParams) error {
	q := New(db.pool)
	_, err := q.UpdateLobbyCountry(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			// TODO:: better error, this could also mean the lobby doesnt exist
			return ErrCountryDoesntExist
		}
		return err
	}
	return nil
}

func (db *Database) GetCountries(ctx context.Context) ([]GetCountriesRow, error) {
	q := New(db.pool)
	countries, err := q.GetCountries(ctx)
	if err != nil {
		return nil, err
	}
	return countries, nil
}
