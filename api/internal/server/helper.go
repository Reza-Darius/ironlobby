package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// helper functions for encoding json bodies
func encode[T any](w http.ResponseWriter, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	w.WriteHeader(status)
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

// GetPlayerID get player id from request context
func GetPlayerID(r *http.Request) uuid.UUID {
	return r.Context().Value(CookieName).(uuid.UUID)
}

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")
		next.ServeHTTP(w, r)
	})
}

// errors if it cant retrieve both lobby id and player id
func getIDs(r *http.Request) (uuid.UUID, int64, error) {
	// player ID from cookie
	playerID := GetPlayerID(r)

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
