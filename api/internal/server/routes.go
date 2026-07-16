package server

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *Application) routes() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(secureHeaders)

	if app.config.DebugCors {
		slog.Info("CORS debug enabled")
		r.Use(CorsDebug)
	}

	r.Route("/api", func(r chi.Router) {
		// unauthorized routes
		r.Get("/health", app.healthcheck)
		r.Get("/countries", app.getCountries)
		r.Get("/lobby/{lobby_id}", app.getLobby)
		r.Get("/lobby/{lobby_id}/country", app.getLobbyCountries)
		r.Get("/lobby/{lobby_id}/player", app.getLobbyPlayers)
		r.Post("/user", app.newUser)

		// authorized routes, require a user name with a corresponding cookie with the user's id
		r.Group(func(r chi.Router) {
			r.Use(app.AuthSession)

			// host actions
			
			r.Post("/lobby", app.newLobby)
			r.Patch("/lobby/{lobby_id}", app.updateLobby)
			r.Post("/lobby/{lobby_id}/country", app.addLobbyCountry)
			r.Delete("/lobby/{lobby_id}/country/{country_tag}", app.deleteLobbyCountry)
			r.Patch("/lobby/{lobby_id}/country/{country_tag}", app.updateLobbyCountry)

			// player actions
			
			// idempotent route, which can also be used for updating the user, like swapping tags
			r.Post("/lobby/{lobby_id}/player", app.joinLobby) 
			r.Delete("/lobby/{lobby_id}/player", app.leaveLobby)
		})
	})

	return r
}
