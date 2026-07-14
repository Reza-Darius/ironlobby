package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
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
		pgErr, _ := errors.AsType[*pgconn.PgError](err)
		if pgErr.Code == pgerrcode.UniqueViolation {
			// handle duplicate
			slog.Debug("new user request with duplicate name", "name", username)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("error when inserting new user into db", "err", err)
		return
	}
	WriteUserCookie(w, id)

	slog.Info("new user registered", "username", username, "id", id)
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
		// TODO: 404 not found for non existant lobby
		w.WriteHeader(http.StatusInternalServerError)
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
		slog.Error("create lobby db insert error", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	slog.Info("new lobby created", "lobby", lobby)
}

func (app *Application) joinLobby(w http.ResponseWriter, r *http.Request) {
	joinParams, err := decode[database.AssignPlayerToLobbyParams](r)
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
		// TODO: error code in case country is occupied
	}

	slog.Info("player joined lobby", "player", joinParams.PlayerID.String(), "lobby", joinParams.LobbyID, "tag", joinParams.CountryTag)
}
