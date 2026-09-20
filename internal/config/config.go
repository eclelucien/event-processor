package config

import (
	"fmt"
	"os"
)

type Config struct {
	AppEnv      string
	HTTPPort    string
	DatabaseURL string
}

func Load() (Config, error) {
	appEnv := os.Getenv("APP_ENV")
	httpPort := os.Getenv("HTTP_PORT")
	databaseURL := os.Getenv("DATABASE_URL")

	if appEnv == "" {
		return Config{}, fmt.Errorf("APP_ENV is not set")
	}

	if httpPort == "" {
		return Config{}, fmt.Errorf("HTTP_PORT is not set")
	}

	if databaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is not set")
	}

	return Config{
		AppEnv:      appEnv,
		HTTPPort:    httpPort,
		DatabaseURL: databaseURL,
	}, nil
}
