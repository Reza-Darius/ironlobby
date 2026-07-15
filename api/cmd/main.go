package main

import (
	"log/slog"
	"os"

	"github.com/reza-darius/ironlobby/internal/database"
	"github.com/reza-darius/ironlobby/internal/server"
	"github.com/reza-darius/ironlobby/internal/utils"
)

func main() {
	utils.InitLogging()

	config, err := utils.LoadConfigEnv()
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	db, err := database.NewDB(config.DBUrl)
	if err != nil {
		slog.Error("database init error")
		os.Exit(1)
	}

	err = server.Run(config.Addr, config.Port, db)
	if err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}
