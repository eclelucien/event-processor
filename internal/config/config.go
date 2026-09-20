package config

import (
	"fmt"
	"os"
)

type Config struct {
	AppEnv   string
	HTTPPort string
}

func Load() (Config, error) {
	appEnv := os.Getenv("APP_ENV")
	httpPort := os.Getenv("HTTP_PORT")

	if appEnv == "" {
		return Config{}, fmt.Errorf("APP_ENV is not set")
	}

	if httpPort == "" {
		return Config{}, fmt.Errorf("HTTP_PORT is not set")
	}

	return Config{
		AppEnv:   appEnv,
		HTTPPort: httpPort,
	}, nil
}
