package server

import (
	"github.com/go-chi/chi/v5"
)

func (app *Application) routes() *chi.Mux {
	r := chi.NewRouter()

	r.Use(secureHeaders)
	r.Route("/api", func(r chi.Router) {
		// unauthorized routes
		r.Get("/health", app.healthcheck)
		r.Get("/{lobby_id}", app.getLobby)
		r.Post("/user", app.newUser)

		// authorized routes
		r.Group(func(r chi.Router) {
			r.Use(app.AuthSession)
			r.Post("/lobby", app.newLobby)
			r.Post("/{lobby_id}", app.joinLobby)
		})
	})

	return r
}
