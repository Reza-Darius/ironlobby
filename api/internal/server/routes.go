package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *Application) routes() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(secureHeaders)

	r.Route("/api", func(r chi.Router) {
		// unauthorized routes
		r.Get("/health", app.healthcheck)
		r.Get("/country", app.getCountries)
		r.Get("/lobby/{lobby_id}", app.getLobby)
		r.Post("/user", app.newUser)

		// authorized routes
		r.Group(func(r chi.Router) {
			r.Use(app.AuthSession)
			r.Post("/lobby", app.newLobby)
			r.Patch("/lobby/{lobby_id}", app.updateLobby)
			r.Post("/lobby/{lobby_id}/country", app.addLobbyCountry)

			// r.Delete("lobby/{lobby_idy}/country/{country_tag}", app.deleteLobbyCountry)
			// r.Put("lobby/{lobby_id}/{country_tag}", app.editLobbyCountry)

			r.Post("/lobby/{lobby_id}/player", app.joinLobby)
			// r.Patch("/lobby/{lobby_id}/player", app.editPlayerSlot)
		})
	})

	return r
}
