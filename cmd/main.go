package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/reza-darius/ironlobby/internal/database"
	"github.com/reza-darius/ironlobby/internal/server"
	"github.com/reza-darius/ironlobby/internal/utils"
)

func main() {
	utils.InitLogging()

	config, err := utils.LoadConfig()
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	_, err = database.NewDB(config.DBUrl)
	if err != nil {
		slog.Error("database init error")
		os.Exit(1)
	}

	routes := server.NewRouter()

	addr := config.Addr + ":" + config.Port
	slog.Info("listening", "addr", addr)

	err = http.ListenAndServe(addr, routes)
	if err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}
