// Package server composes the primary iron lobby server
package server

import (
	"log/slog"
	"net/http"

	"github.com/reza-darius/ironlobby/internal/database"
)

type Application struct {
	db *database.Database
}

func Run(addr string, port string, db *database.Database) error {
	app := Application {
		db: db,
	}

	a := addr + ":" + port
	slog.Info("listening", "addr", a)

	err := http.ListenAndServe(a, app.routes())
	if err != nil {
		slog.Error("server error", "err", err)
		return err
	}
	return nil
}
