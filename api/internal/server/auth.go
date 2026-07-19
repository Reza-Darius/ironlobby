package server

import (
	"context"
	"log"
	"log/slog"
	"math"
	"net/http"
	"github.com/google/uuid"
)

const UserIDCookie = "userID"

func (app *Application) authSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		{
			cookie, err := r.Cookie(UserIDCookie)
			if err != nil || cookie.Value == "" {
				if err == http.ErrNoCookie {
					log.Println("no cookie detected, redirecting")

					// frontend handles redirect
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				slog.Error("auth handler error when retrieving cookie", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			intUUID, err := uuid.Parse(cookie.Value)
			if err != nil {
				slog.Error("UUID could not be parsed", "err", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// validate user exists
			// OPTIMIZE: user id cache instead of DB queries
			_, err = app.db.GetUser(r.Context(), intUUID)
			if err != nil {
				slog.Error("could not retrieve user from db", "err", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// create new context with attached key value pair, inherit parent context (request)
			ctx := context.WithValue(r.Context(), UserIDCookie, intUUID)

			// attach context to next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

func WriteUserCookie(w http.ResponseWriter, userID uuid.UUID, debug bool) {
	cookie :=  &http.Cookie{
		Name:     UserIDCookie,
		Value:    userID.String(),
		Path:     "/",
		MaxAge:   math.MaxInt32, // cookie is effectively permanent
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	if debug {
		cookie.Secure = false
		cookie.HttpOnly = false
	}
	http.SetCookie(w, cookie)
	w.Header().Add("Vary", "Cookie")
	w.Header().Add("Cache-Control", `no-cache="Set-Cookie"`)
}

