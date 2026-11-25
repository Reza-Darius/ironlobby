package server

import (
	"log/slog"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const TIMEOUT = time.Second * 20

func (app *Application) routes() *chi.Mux {
	r := chi.NewRouter()

	// global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(secureHeaders)
	r.Use(timeoutHandler(TIMEOUT))
	if app.config.DebugCors {
		slog.Info("CORS debug enabled")
		r.Use(CorsDebug)
	}

	r.Route("/api", func(r chi.Router) {
		// unauthorized routes
		r.Post("/user", app.newUser)

		r.Get("/health", app.healthcheck)
		r.Get("/countries", app.getCountries)

		r.Get("/lobby/{lobby_id}", app.getLobby)
		r.Get("/lobby/{lobby_id}/country", app.getLobbyCountries)
		r.Get("/lobby/{lobby_id}/player", app.getLobbyPlayers)

		// authorized routes, require a user name with a corresponding cookie with the user's id
		r.Group(func(r chi.Router) {
			r.Use(app.AuthSession)

			// player actions
			r.Post("/lobby", app.newLobby)

			// idempotent route, which can also be used for updating the user, like swapping tags
			r.Post("/lobby/{lobby_id}/player", app.joinLobby)
			r.Delete("/lobby/{lobby_id}/player", app.leaveLobby)

			// host actions
			r.Group(func(r chi.Router) {
				r.Use(app.AuthHost)

				r.Patch("/lobby/{lobby_id}", app.updateLobby)
				r.Delete("/lobby/{lobby_id}", app.deleteLobby)
				r.Post("/lobby/{lobby_id}/country", app.addLobbyCountry)
				r.Delete("/lobby/{lobby_id}/country/{country_tag}", app.deleteLobbyCountry)
				r.Patch("/lobby/{lobby_id}/country/{country_tag}", app.updateLobbyCountry)
			})
		})
	})

	return r
}
