package server

import (
	"github.com/go-chi/chi/v5"
)

func (app *Application) routes() *chi.Mux {
	r := chi.NewRouter()
	r.Get("/health", app.healthcheck)

	// unauthorized routes
	r.Get("/{lobby_id}", app.getLobby)
	r.Post("/user", app.newUser)

	// authorized routes
	r.Post("/lobby", app.newLobby)

	return r
}
