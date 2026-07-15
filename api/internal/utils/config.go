package utils

import (
	"log/slog"

	"github.com/ilyakaznacheev/cleanenv"
)

type AppConfig struct {
	Addr          string `env:"ADDR" env-default:"0.0.0.0"`
	Port          string `env:"PORT" env-default:"80"`
	DBUrl         string `env:"DATABASE_URL"`
	DebugCors    bool   `env:"DEBUG_CORS" env-default:"false"`
}

func LoadConfigEnv() (AppConfig, error) {
	var cfg AppConfig
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		return AppConfig{}, err
	}
	slog.Info("config loaded from env variables", "config", cfg)
	return cfg, nil
}

func LoadConfigFile(path string) (AppConfig, error) {
	var cfg AppConfig
	err := cleanenv.ReadConfig(path, &cfg)
	if err != nil {
		return AppConfig{}, err
	}
	slog.Info("config loaded from file", "path", path, "config", cfg)
	return cfg, nil
}
