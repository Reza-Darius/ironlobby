package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// JSON writes data as a JSON response with the given HTTP status code.
// It sets the Content-Type header to application/json. If marshalling fails,
// a 500 Internal Server Error is written as plain text.
func respondJSON(w http.ResponseWriter, status int, data any) {
	body, err := json.Marshal(data)
	if err != nil {
		slog.Error("failed to marshal json", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(body)
}

// helper functions for decoding json bodies
func decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
}

// GetPlayerID get player id from request context
func GetPlayerID(r *http.Request) uuid.UUID {
	return r.Context().Value(CookieName).(uuid.UUID)
}

// retrieves player and lobby id, errors if it cant retrieve both lobby id and player id
func getIDs(r *http.Request) (uuid.UUID, int64, error) {
	// player ID from cookie
	playerID := GetPlayerID(r)

	// lobby ID from url path
	lobbyID, err := getLobbyID(r)
	if err != nil {
		return uuid.UUID{}, 0, err
	}

	return playerID, lobbyID, nil
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

func getPlayerNameFromPath(r *http.Request) (string, error) {
	playerName := chi.URLParam(r, "player_name")
	if playerName == "" {
		return "", errors.New("empty play name")
	}
	unescapedName, err := url.PathUnescape(playerName) 
	if err != nil {
		return "", errors.New("invalid name: path escapes")
	}
	return unescapedName, nil
}
