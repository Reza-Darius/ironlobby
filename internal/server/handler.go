package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
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

type NewUserRequest struct {
	Username string `json:"username"`
}

func (app *Application) newUser(w http.ResponseWriter, r *http.Request) {
	username, err := decode[NewUserRequest](r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("error when decoding request body for new user", "err", err)
		return
	}

	id, err := app.db.NewUser(r.Context(), username.Username)
	if err != nil {
		pgErr,_ := errors.AsType[*pgconn.PgError](err)
    if pgErr.Code == pgerrcode.UniqueViolation {
			// handle duplicate
			w.WriteHeader(http.StatusBadRequest)
			return
    }
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("error when inserting new user into db", "err", err)
		return
	}

	WriteUserCookie(w, id)
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
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("error when fetching lobby from db", "err", err)
		return
	}

	err = encode(w, r, http.StatusOK, lobby)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("error when encoding json body in get lobby", "err", err)
		return
	}
}

func (app *Application) newLobby(w http.ResponseWriter, r *http.Request) {
	username, err := decode[NewUserRequest](r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("error when decoding request body for new user", "err", err)
		return
	}

	id, err := app.db.NewUser(r.Context(), username.Username)
	if err != nil {
		pgErr,_ := errors.AsType[*pgconn.PgError](err)
    if pgErr.Code == pgerrcode.UniqueViolation {
			// handle duplicate
			w.WriteHeader(http.StatusBadRequest)
			return
    }
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("error when inserting new user into db", "err", err)
		return
	}

	WriteUserCookie(w, id)
}

type JoinLobbyRequest struct {
	Country string `json:"country_tag"`
}

func (app *Application) joinLobby(w http.ResponseWriter, r *http.Request) {
	lobbyID := chi.URLParam(r, "lobby_id")
	if lobbyID == "" {
		w.WriteHeader(http.StatusBadRequest)
		slog.Error("invalid lobby ID provided")
		return
	}

	lobbyIDint, err := strconv.ParseInt(lobbyID, 10, 64)
	if err != nil {
		slog.Error("lobby int parse error", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return 
	}

	country, err := decode[JoinLobbyRequest](r)
	if err != nil {
		slog.Error("join lobby request body decode error", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return 
	}

	playerID := GetPlayerID(r)

	err = app.db.JoinLobby(r.Context(), lobbyIDint, playerID, country.Country)
	if err != nil {
		// TODO: error code in case country is occupied
	}
}
