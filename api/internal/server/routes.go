package server

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *Application) routes() *chi.Mux {
	r := chi.NewRouter()

	// for colored logging in docker
	middleware.IsTTY = true

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Compress(5))
	r.Use(secureHeaders)

	if app.config.DebugCors {
		slog.Info("CORS debug enabled")
		r.Use(corsDebug)
	}

	r.Route("/api", func(r chi.Router) {
		// unauthorized routes
		r.Post("/user", app.newUser)

		r.Get("/health", app.healthcheck)
		r.Get("/countries", app.getCountries)

		r.Route("/lobby/{lobby_id}", func(r chi.Router) {
			r.Get("/", app.getLobby)
			r.Get("/country", app.getLobbyCountries)
			r.Get("/player", app.getLobbyPlayers)
		})

		// authorized routes, require a user name with a corresponding cookie with the user's id
		r.Group(func(r chi.Router) {
			r.Use(app.AuthSession)

			r.Route("/lobby", func(r chi.Router) {
				// host actions

				r.Post("/", app.newLobby)
				r.Patch("/{lobby_id}", app.updateLobby)

				r.Post("/{lobby_id}/country", app.addLobbyCountry)
				r.Delete("/{lobby_id}/country/{country_tag}", app.deleteLobbyCountry)
				r.Patch("/{lobby_id}/country/{country_tag}", app.updateLobbyCountry)

				// player actions

				// idempotent route, which can also be used for updating the user, like swapping tags
				r.Post("/{lobby_id}/player", app.joinLobby)
				r.Delete("/{lobby_id}/player", app.leaveLobby)
			})
		})
	})

	return r
}
