package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

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
