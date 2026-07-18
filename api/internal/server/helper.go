package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/reza-darius/ironlobby/internal/database"
)

// helper functions for encoding json bodies
func encode[T any](w http.ResponseWriter, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

// helper functions for decoding json bodies
func decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
}


// errors if it cant retrieve both lobby id and player id
func getIDs(r *http.Request) (uuid.UUID, int64, error) {
	// player ID from cookie
	playerID := getUserID(r)

	// lobby ID from url path
	lID := chi.URLParam(r, "lobby_id")
	if lID == "" {
		return uuid.UUID{}, 0, errors.New("empty lobby ID")
	}

	lobbyID, err := strconv.ParseInt(lID, 10, 64)
	if err != nil {
		return uuid.UUID{}, 0, err
	}
	return playerID, lobbyID, nil
}

// getUserID get player id from request context
func getUserID(r *http.Request) uuid.UUID {
	return r.Context().Value(UserIDCookie).(uuid.UUID)
}

func getLobbyID(r *http.Request) (int64, error) {
	lobbyID := chi.URLParam(r, "lobby_id")
	if lobbyID == "" {
		return 0, errors.New("empty lobby ID")
	}

	IDInt, err := strconv.ParseInt(lobbyID, 10, 64)
	if err != nil {
		return 0, err
	}
	return IDInt, nil
}

// returns IDs if the user id is host for the lobby id
func checkHost(db *database.Database, w http.ResponseWriter, r *http.Request) (uuid.UUID, int64, error) {
	playerID, lobbyID, err := getIDs(r)
	if err != nil {
		slog.Error("couldnt retrieve lobby or player id", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return uuid.UUID{}, 0, fmt.Errorf("couldnt get IDs, err: %v", err)
	}

	// check host privileges
	isHost, err := db.PlayerIsHost(r.Context(), playerID, lobbyID)
	if err != nil {
		slog.Error("couldnt check host in db", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return uuid.UUID{}, 0, fmt.Errorf("couldnt check DB for host, err: %v", err)
	}

	if !isHost {
		slog.Error("user is not host")
		w.WriteHeader(http.StatusUnauthorized)
		return uuid.UUID{}, 0, fmt.Errorf("request user is not host, err: %v", err)
	}
	return playerID, lobbyID, nil
}
