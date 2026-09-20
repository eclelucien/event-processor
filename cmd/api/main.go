package main

import (
	"log"

	"github.com/eclesiaste/event-processor/internal/config"
	httpserver "github.com/eclesiaste/event-processor/internal/interfaces/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	server := httpserver.NewServer(cfg)

	log.Printf(
		"Event Processor running on port %s (%s)",
		cfg.HTTPPort,
		cfg.AppEnv,
	)

	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
