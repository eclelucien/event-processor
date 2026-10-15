package config

import (
	"fmt"
	"os"
)

type Config struct {
	AppEnv       string
	HTTPPort     string
	DatabaseURL  string
	KafkaBrokers string
	KafkaTopic   string
	KafkaGroupID string
}

func Load() (Config, error) {
	appEnv := os.Getenv("APP_ENV")
	httpPort := os.Getenv("HTTP_PORT")
	databaseURL := os.Getenv("DATABASE_URL")
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	kafkaGroupID := os.Getenv("KAFKA_GROUP_ID")

	if appEnv == "" {
		return Config{}, fmt.Errorf("APP_ENV is not set")
	}

	if httpPort == "" {
		return Config{}, fmt.Errorf("HTTP_PORT is not set")
	}

	if databaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is not set")
	}

	if kafkaBrokers == "" {
		return Config{}, fmt.Errorf("KAFKA_BROKERS is not set")
	}

	if kafkaTopic == "" {
		return Config{}, fmt.Errorf("KAFKA_TOPIC is not set")
	}

	if kafkaGroupID == "" {
		return Config{}, fmt.Errorf("KAFKA_GROUP_ID is not set")
	}

	return Config{
		AppEnv:       appEnv,
		HTTPPort:     httpPort,
		DatabaseURL:  databaseURL,
		KafkaBrokers: kafkaBrokers,
		KafkaTopic:   kafkaTopic,
		KafkaGroupID: kafkaGroupID,
	}, nil
}
