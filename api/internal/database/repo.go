package database

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/reza-darius/ironlobby/internal/database/sqlc"
)

// domain errors to not expose PG internals
var (
	ErrPlayerNotFound = errors.New("user not found")

	ErrLobbyExists       = errors.New("lobby already exists")
	ErrLobbyDoesntExists = errors.New("lobby doesnt exist")

	ErrUserExists            = errors.New("user already exists")
	ErrNationSlotsFull       = errors.New("the requested nation's slots are full")
	ErrNationNotAvail        = errors.New("the requested nation is not available")
	ErrPlayerAlreadyAssigned = errors.New("player is already assigned to Nation")
	ErrCountryDoesntExist    = errors.New("provided country tag doesnt exist")
)

func (db *Database) OpenLobbies(ctx context.Context) (int64, error) {
	q := sqlc.New(db.pool)
	return q.OpenLobbies(ctx)
}

func (db *Database) GetPlayer(ctx context.Context, id uuid.UUID) (string, error) {
	q := sqlc.New(db.pool)
	name, err := q.GetPlayer(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", ErrPlayerNotFound
		}
		return "", err
	}
	return name, nil
}

func (db *Database) GetPlayerID(ctx context.Context, playerName string) (uuid.UUID, error) {
	q := sqlc.New(db.pool)
	id, err := q.GetPlayerID(ctx, playerName)
	if err != nil {
		if err == pgx.ErrNoRows {
			return uuid.UUID{}, ErrPlayerNotFound
		}
		return uuid.UUID{}, err
	}
	return id, nil
}

func (db *Database) NewUser(ctx context.Context, playerName string) (uuid.UUID, error) {
	q := sqlc.New(db.pool)
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
	Lobby     sqlc.InsertLobbyParams          `json:"lobby"`
	Countries []sqlc.UpsertLobbyCountryParams `json:"countries"`
}

func (db *Database) NewLobby(ctx context.Context, arg CreateLobbyRequest) (sqlc.Lobby, error) {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return sqlc.Lobby{}, err
	}
	defer tx.Rollback(ctx)

	q := sqlc.New(db.pool).WithTx(tx)

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
		return sqlc.Lobby{}, err
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
			return sqlc.Lobby{}, err
		}
	}
	if err := br.Close(); err != nil {
		return sqlc.Lobby{}, err
	}
	return lobby, tx.Commit(ctx)
}

// UpsertLobbyPlayer adds a player to a lobby or changes the player's country tag inside the lobby
func (db *Database) UpsertLobbyPlayer(ctx context.Context, arg sqlc.UpsertLobbyPlayerParams) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := sqlc.New(db.pool).WithTx(tx)

	// we lock the targeted country tag
	maxSlots, err := q.LockLobbyCountrySlot(ctx, sqlc.LockLobbyCountrySlotParams{
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
	numOccuppied, err := q.CountCountryOccupants(ctx, sqlc.CountCountryOccupantsParams{
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

func (db *Database) DeleteLobbyPlayer(ctx context.Context, arg sqlc.DeleteLobbyPlayerParams) error {
	q := sqlc.New(db.pool)
	_, err := q.DeleteLobbyPlayer(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ErrPlayerNotFound
		}
		return err
	}
	return nil
}

type LobbyInfo struct {
	Lobby     sqlc.Lobby                  `json:"lobby"`
	Countries []sqlc.GetLobbyCountriesRow `json:"countries"`
	Players   []sqlc.GetLobbyPlayersRow   `json:"players"`
}

func (db *Database) GetLobby(ctx context.Context, lobbyID int64) (LobbyInfo, error) {
	q := sqlc.New(db.pool)
	lobby, err := q.GetLobby(ctx, lobbyID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return LobbyInfo{}, ErrLobbyDoesntExists
		}
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

func (db *Database) GetLobbyCountries(ctx context.Context, lobbyID int64) ([]sqlc.GetLobbyCountriesRow, error) {
	q := sqlc.New(db.pool)
	lobbyCountries, err := q.GetLobbyCountries(ctx, lobbyID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrLobbyDoesntExists
		}
		return nil, err
	}
	return lobbyCountries, nil
}

func (db *Database) GetLobbyPlayer(ctx context.Context, arg sqlc.GetLobbyPlayerParams) (sqlc.GetLobbyPlayerRow, error) {
	q := sqlc.New(db.pool)
	lobbyPlayer, err := q.GetLobbyPlayer(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			return sqlc.GetLobbyPlayerRow{}, ErrPlayerNotFound
		}
		return sqlc.GetLobbyPlayerRow{}, err
	}
	return lobbyPlayer, nil
}

func (db *Database) GetLobbyPlayers(ctx context.Context, lobbyID int64) ([]sqlc.GetLobbyPlayersRow, error) {
	q := sqlc.New(db.pool)
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
	q := sqlc.New(db.pool)
	_, err := q.GetLobbyFromHostID(ctx, sqlc.GetLobbyFromHostIDParams{
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
func (db *Database) AddLobbyCountry(ctx context.Context, arg sqlc.UpsertLobbyCountryParams) error {
	q := sqlc.New(db.pool)
	_, err := q.UpsertLobbyCountry(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ErrCountryDoesntExist
		}
		return err
	}
	return nil
}

func (db *Database) DeleteLobbyCountry(ctx context.Context, arg sqlc.DeleteLobbyCountryParams) error {
	q := sqlc.New(db.pool)
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

func (db *Database) UpdateLobby(ctx context.Context, arg sqlc.UpdateLobbyParams) error {
	q := sqlc.New(db.pool)
	_, err := q.UpdateLobby(ctx, arg)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) DeleteLobby(ctx context.Context, arg int64) error {
	q := sqlc.New(db.pool)
	affected, err := q.DeleteLobby(ctx, arg)
	if affected == 0 {
		return ErrLobbyDoesntExists
	}

	if err != nil {
		return err
	}
	return nil
}

func (db *Database) UpdateLobbyCountry(ctx context.Context, arg sqlc.UpdateLobbyCountryParams) error {
	q := sqlc.New(db.pool)
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

func (db *Database) GetCountries(ctx context.Context) ([]sqlc.GetCountriesRow, error) {
	q := sqlc.New(db.pool)
	countries, err := q.GetCountries(ctx)
	if err != nil {
		return nil, err
	}
	return countries, nil
}

func (db *Database) TagIsValid(ctx context.Context, tag string) (bool, error) {
	if len(tag) != 3 { // every tag is 3 letters
		return false, nil
	}
	tag = strings.ToUpper(tag)
	q := sqlc.New(db.pool)
	matches, err := q.TagExists(ctx, tag)
	if err != nil {
		return false, err
	}
	return matches == 1, nil
}
