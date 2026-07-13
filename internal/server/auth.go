package server

import (
	"context"
	"log"
	"log/slog"
	"math"
	"net/http"
	"github.com/google/uuid"
)

const CookieName = "username"

func (app *Application) AuthSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		{
			cookie, err := r.Cookie(CookieName)
			if err != nil || cookie.Value == "" {
				// redirect to login page
				if err == http.ErrNoCookie {
					log.Println("no cookie detected, redirecting")

					http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
					return
				}
				slog.Error("auth handler error when retrieving cookie", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// err = app.Sessions.Authenticate(r)
			// if err != nil {
			// 	if err == auth.SessionExpired {
			// 		log.Println("cookies expired")
			// 		http.Redirect(w, r, "/login", 307)
			// 		return
			// 	}
			// 	slog.Info("authentication failed")
			// 	w.WriteHeader(http.StatusUnauthorized)
			// 	w.Write([]byte("unauthorized"))
			// 	return
			// }

			// create new context with attached key value pair, inherit parent context (request)
			ctx := context.WithValue(r.Context(), "userID", cookie.Value)

			// attach context to next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

func WriteUserCookie(w http.ResponseWriter, userID uuid.UUID) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    userID.String(),
		Path:     "/",
		MaxAge:   math.MaxInt32, // cookie is effectively permanent
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Add("Vary", "Cookie")
	w.Header().Add("Cache-Control", `no-cache="Set-Cookie"`)
}
