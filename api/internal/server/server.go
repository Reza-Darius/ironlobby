// Package server composes the primary iron lobby server
package server

import (
	"log/slog"
	"net/http"

	"github.com/reza-darius/ironlobby/internal/database"
	"github.com/reza-darius/ironlobby/internal/utils"
)

type Application struct {
	db *database.Database
	config *utils.AppConfig
}

func Run(config *utils.AppConfig, db *database.Database) error {
	app := Application {
		db: db,
	}

	a := config.Addr + ":" + config.Port
	slog.Info("listening", "addr", a)

	err := http.ListenAndServe(a, app.routes())
	if err != nil {
		slog.Error("server error", "err", err)
		return err
	}
	return nil
}
