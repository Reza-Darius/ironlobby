package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/reza-darius/ironlobby/internal/database"
)

func (app *Application) healthcheck(w http.ResponseWriter, r *http.Request) {
	openLobbies, err := app.db.OpenLobbies(r.Context())
	if err != nil {
		slog.Error("health check error when fetching open lobbies from db", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = fmt.Fprintf(w, "open lobbies: %v", openLobbies)
	if err != nil {
		slog.Error("health check error when writing to response", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (app *Application) newUser(w http.ResponseWriter, r *http.Request) {
	var username struct {
		Username string
	}

	if err := json.NewDecoder(r.Body).Decode(&username); err != nil {
		slog.Error("error when decoding request body for new user", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := app.db.NewUser(r.Context(), username.Username)
	if err != nil {
		switch err {

		case database.ErrUserExists:
			{
				http.Error(w, "user already exists", http.StatusConflict)
			}

		default:
			{

				slog.Error("error when inserting new user into db", "err", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}
		return
	}
	WriteUserCookie(w, id, app.config.DebugCors)

	w.WriteHeader(http.StatusCreated)
	slog.Debug("new user registered", "username", username, "id", id)
}

func (app *Application) getCountries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	countries, err := app.db.GetCountries(ctx)
	if err != nil {
		http.Error(w, "couldnt fetch countries", http.StatusInternalServerError)
		return
	}
	err = encode(w, http.StatusOK, countries)
	if err != nil {
		http.Error(w, "couldnt encode countries", http.StatusInternalServerError)
		return
	}
}

func (app *Application) getLobby(w http.ResponseWriter, r *http.Request) {
	lobbyID, err := getLobbyID(r)
	if err != nil {
		http.Error(w, "couldnt retrieve lobby id", http.StatusBadRequest)
		return
	}

	lobby, err := app.db.GetLobbyInfo(r.Context(), lobbyID)
	if err != nil {
		switch err {
		case database.ErrLobbyDoesntExists:
			{
				http.Error(w, "lobby not found", http.StatusNotFound)
			}

		default:
			{
				slog.Error("failed to get lobby", "err", err, "lobby", lobbyID)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}
		return
	}

	err = encode(w, http.StatusOK, lobby)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("error when encoding json body in get lobby", "err", err)
		return
	}
}

func (app *Application) getLobbyCountries(w http.ResponseWriter, r *http.Request) {
	lobbyID, err := getLobbyID(r)
	if err != nil {
		http.Error(w, "couldnt retrieve lobby id", http.StatusBadRequest)
		return
	}

	lobby, err := app.db.GetLobbyCountries(r.Context(), lobbyID)
	if err != nil {
		switch err {
		case database.ErrLobbyDoesntExists:
			{
				http.Error(w, "lobby not found", http.StatusNotFound)
			}

		default:
			{
				slog.Error("error when fetching lobby countries", "err", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}
		return
	}

	err = encode(w, http.StatusOK, lobby)
	if err != nil {
		slog.Error("error when encoding json body in get lobby countries", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (app *Application) getLobbyPlayers(w http.ResponseWriter, r *http.Request) {
	lobbyID, err := getLobbyID(r)
	if err != nil {
		http.Error(w, "couldnt retrieve lobby id", http.StatusBadRequest)
		return
	}

	lobby, err := app.db.GetLobbyPlayers(r.Context(), lobbyID)
	if err != nil {
		switch err {
		case database.ErrLobbyDoesntExists:
			{
				http.Error(w, "lobby not found", http.StatusNotFound)
			}

		default:
			{
				slog.Error("error when fetching lobby countries", "err", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}
		return
	}

	err = encode(w, http.StatusOK, lobby)
	if err != nil {
		slog.Error("error when encoding json body in get lobby player", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (app *Application) newLobby(w http.ResponseWriter, r *http.Request) {
	lobbyParams, err := decode[database.CreateLobbyRequest](r)
	if err != nil {
		slog.Error("create lobby request body decode error", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// we get the hostID from the cookie
	lobbyParams.Lobby.HostPlayer = getUserID(r)

	lobby, err := app.db.NewLobby(r.Context(), lobbyParams)
	if err != nil {
		switch err {

		case database.ErrLobbyExists:
			{
				http.Error(w, "lobby already exists", http.StatusConflict)
			}

		default:
			{
				slog.Error("error when creating lobby", "err", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}
		return
	}

	newLobbyRes := struct {
		LobbyID int64 `json:"lobby_id"`
	}{
		LobbyID: lobby.ID,
	}

	err = encode(w, http.StatusCreated, newLobbyRes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("error when encoding json body in create lobby", "err", err)
		return
	}

	w.Header().Add("Location", "/lobby/"+strconv.FormatInt(lobby.ID, 10))
	slog.Debug("new lobby created", "lobby", lobby)
}

func (app *Application) joinLobby(w http.ResponseWriter, r *http.Request) {
	joinParams, err := decode[database.UpsertLobbyPlayerParams](r)
	if err != nil {
		slog.Error("join lobby request body decode error", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	playerID, lobbyID, err := getIDs(r)
	if err != nil {
		http.Error(w, "invalid ids provided", http.StatusBadRequest)
		return
	}

	joinParams.PlayerID = playerID
	joinParams.LobbyID = lobbyID

	err = app.db.UpsertLobbyPlayer(r.Context(), joinParams)
	if err != nil {
		switch err {
		case database.ErrNationSlotsFull:
			{
				http.Error(w, "requested nation is full", http.StatusBadRequest)
				return
			}
		case database.ErrNationNotAvail:
			{
				http.Error(w, "requested nation is unavailable", http.StatusBadRequest)
				return
			}

		default:
			{
				slog.Error("error when creating lobby", "err", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	slog.Debug("player joined lobby", "player", joinParams.PlayerID.String(), "lobby", joinParams.LobbyID, "tag", joinParams.CountryTag)
}

func (app *Application) leaveLobby(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	playerID, lobbyID, err := getIDs(r)
	if err != nil {
		http.Error(w, "invalid ids provided", http.StatusBadRequest)
		return
	}
	err = app.db.DeleteLobbyPlayer(ctx, database.DeleteLobbyPlayerParams{
		PlayerID: playerID,
		LobbyID:  lobbyID,
	})
	if err != nil {
		switch err {
		case database.ErrUserNotFound:
			{
				http.Error(w, "requested player doesnt exist in lobby or lobby doesnt exist", http.StatusNotFound)
				return
			}

		default:
			{
				slog.Error("error when updating lobby country", "err", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
	slog.Debug("deleted user from lobby", "player", playerID, "lobby", lobbyID)
}

func (app *Application) updateLobby(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	_, lobbyID, err := checkHost(app.db, w, r)
	if err != nil {
		return
	}

	updateParams, err := decode[database.UpdateLobbyParams](r)
	if err != nil {
		slog.Error("update lobby request body decode error", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	updateParams.LobbyID = lobbyID

	err = app.db.UpdateLobby(ctx, updateParams)
	if err != nil {
		slog.Error("failed to run update lobby on db", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	slog.Debug("updated lobby", "lobby", lobbyID, "lobby_settings", updateParams)
}

func (app *Application) addLobbyCountry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	_, lobbyID, err := checkHost(app.db, w, r)
	if err != nil {
		return
	}

	// parse body
	addLobbyCountryParams, err := decode[database.UpsertLobbyCountryParams](r)
	if err != nil {
		slog.Error("add lobby country request body decode error", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	addLobbyCountryParams.LobbyID = lobbyID

	err = app.db.AddLobbyCountry(ctx, addLobbyCountryParams)
	if err != nil {
		switch err {
		case database.ErrCountryDoesntExist:
			{
				http.Error(w, "requested country doesnt exist", http.StatusNotFound)
				return
			}

		default:
			{
				slog.Error("error when adding lobby to country", "err", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}
		return
	}
	w.WriteHeader(http.StatusCreated)
	slog.Debug("added country to lobby", "lobby", lobbyID, "country", addLobbyCountryParams.CountryTag, "max_slots", addLobbyCountryParams.MaxSlots)
}

func (app *Application) deleteLobbyCountry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, lobbyID, err := checkHost(app.db, w, r)
	if err != nil {
		slog.Error("user is not host")
		return
	}

	tag := chi.URLParam(r, "country_tag")
	tag = strings.ToUpper(tag)
	if tag == "" {
		slog.Debug("invalid country tag", "provided", tag)
		http.Error(w, "invalid country tag", http.StatusBadRequest)
		return
	}

	err = app.db.DeleteLobbyCountry(ctx, database.DeleteLobbyCountryParams{
		LobbyID:    lobbyID,
		CountryTag: tag,
	})
	if err != nil {
		switch err {
		case database.ErrCountryDoesntExist:
			{
				http.Error(w, "requested country or lobby doesnt exist", http.StatusNotFound)
				return
			}

		default:
			{
				slog.Error("error when deleting lobby country", "err", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
	slog.Debug("deleted country from lobby", "country", tag, "lobby", lobbyID)
}

func (app *Application) updateLobbyCountry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, lobbyID, err := checkHost(app.db, w, r)
	if err != nil {
		slog.Error("user is not host")
		return
	}

	tag := chi.URLParam(r, "country_tag")
	tag = strings.ToUpper(tag)
	if tag == "" {
		slog.Debug("invalid country tag", "provided", tag)
		http.Error(w, "invalid country tag", http.StatusBadRequest)
		return
	}

	args, err := decode[database.UpdateLobbyCountryParams](r)
	if err != nil {
		slog.Error("delet lobby country request body decode error", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	args.LobbyID = lobbyID
	args.CountryTag = tag

	err = app.db.UpdateLobbyCountry(ctx, args)
	if err != nil {
		switch err {
		case database.ErrCountryDoesntExist:
			{
				http.Error(w, "requested country or lobby doesnt exist", http.StatusNotFound)
				return
			}

		default:
			{
				slog.Error("error when updating lobby country", "err", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}
		return
	}
	slog.Debug("updated country in lobby", "country", tag, "lobby", lobbyID)
}
