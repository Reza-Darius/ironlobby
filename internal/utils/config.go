package utils

import (
	"log/slog"

	"github.com/ilyakaznacheev/cleanenv"
)

type AppConfig struct {
	Addr          string `env:"ADDR" env-default:"0.0.0.0"`
	Port          string `env:"PORT" env-default:"80"`
	MigrationPath string `env:"MIGRATION_PATH"`
	DBUrl         string `env:"DATABASE_URL"`
}

func LoadConfig() (AppConfig, error) {
	var cfg AppConfig
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		return AppConfig{}, err
	}
	slog.Info("config loaded")
	return cfg, nil
}
