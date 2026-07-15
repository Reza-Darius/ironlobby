package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/reza-darius/ironlobby/internal/database"
)

func (app *Application) healthcheck(w http.ResponseWriter, r *http.Request) {
	openLobbies, err := app.db.OpenLobbies(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("health check error when fetching open lobbies from db", "err", err)
		return
	}
	_, err = fmt.Fprintf(w, "open lobbies: %v", openLobbies)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("health check error when writing to response", "err", err)
		return
	}
}

func (app *Application) newUser(w http.ResponseWriter, r *http.Request) {
	var username struct {
		Username string
	}

	if err := json.NewDecoder(r.Body).Decode(&username); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("error when decoding request body for new user", "err", err)
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
	WriteUserCookie(w, id, app.config.Debug_cors)

	slog.Info("new user registered", "username", username, "id", id)
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
	lobbyID := chi.URLParam(r, "lobby_id")
	if lobbyID == "" {
		w.WriteHeader(http.StatusBadRequest)
		slog.Error("invalid lobby ID provided")
		return
	}

	IDInt, err := strconv.ParseInt(lobbyID, 10, 64)
	if err != nil {
		slog.Error("lobby int parse error", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	lobby, err := app.db.GetLobby(r.Context(), IDInt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "lobby not found", http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		slog.Error("error when fetching lobby from db", "err", err)
		return
	}

	err = encode(w, http.StatusOK, lobby)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("error when encoding json body in get lobby", "err", err)
		return
	}
}

func (app *Application) newLobby(w http.ResponseWriter, r *http.Request) {
	lobbyParams, err := decode[database.InsertLobbyParams](r)
	if err != nil {
		slog.Error("create lobby request body decode error", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// we get the hostID from the cookie
	lobbyParams.HostPlayer = GetPlayerID(r)

	lobby, err := app.db.CreateLobby(r.Context(), lobbyParams)
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

	err = encode(w, http.StatusOK, lobby.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("error when encoding json body in create lobby", "err", err)
		return
	}

	slog.Info("new lobby created", "lobby", lobby)
}

func (app *Application) joinLobby(w http.ResponseWriter, r *http.Request) {
	joinParams, err := decode[database.UpsertPlayerLobbyParams](r)
	if err != nil {
		slog.Error("join lobby request body decode error", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// player ID from cookie
	joinParams.PlayerID = GetPlayerID(r)

	// lobby ID from url path
	lobbyID := chi.URLParam(r, "lobby_id")
	if lobbyID == "" {
		w.WriteHeader(http.StatusBadRequest)
		slog.Error("invalid lobby ID provided")
		return
	}

	joinParams.LobbyID, err = strconv.ParseInt(lobbyID, 10, 64)
	if err != nil {
		slog.Error("lobby int parse error", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = app.db.JoinLobby(r.Context(), joinParams)
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

	slog.Info("player joined lobby", "player", joinParams.PlayerID.String(), "lobby", joinParams.LobbyID, "tag", joinParams.CountryTag)
}

func (app *Application) updateLobby(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context();

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

	slog.Info("updated lobby", "lobby", lobbyID, "lobby_settings", updateParams)
}
func (app *Application) addLobbyCountry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context();


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
				http.Error(w, "requested country doesnt exist", http.StatusBadRequest)
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
	slog.Info("added country to lobby", "lobby_id", lobbyID, "country", addLobbyCountryParams.CountryTag, "max_slots", addLobbyCountryParams.MaxSlots)
}

func (app *Application) deleteLobbyCountry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context();
	_, lobbyID, err := checkHost(app.db, w, r)
	if err != nil {
		slog.Error("user is not host")
		return
	}

	tag := chi.URLParam(r, "country_tag")
	tag = strings.ToUpper(tag)
	if tag == "" {
		slog.Error("invalid country tag", "provided", tag)
		http.Error(w, "invalid country tag", http.StatusBadRequest)
		return
	}

	err = app.db.DeleteLobbyCountry(ctx, database.DeleteLobbyCountryParams{
		LobbyID: lobbyID,
		CountryTag: tag,
	})

	if err != nil {
		slog.Error("error when deleting lobby country", "err", err)
		http.Error(w, "unable to fulfill write call", http.StatusInternalServerError)
		return
	}
	slog.Info("deleted country from lobby", "countr", tag, "lobby_id", lobbyID)
}
