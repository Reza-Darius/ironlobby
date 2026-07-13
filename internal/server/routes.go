package server

import (
	"github.com/go-chi/chi/v5"
)

func (app *Application) routes() *chi.Mux {
	r := chi.NewRouter()
	r.Get("/health", app.healthcheck)
	r.Post("/user", app.newUser)

	return r
}
