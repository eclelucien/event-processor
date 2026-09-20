package main

import (
	"context"
	"log"

	"github.com/eclesiaste/event-processor/internal/application"
	"github.com/eclesiaste/event-processor/internal/config"
	"github.com/eclesiaste/event-processor/internal/infrastructure/postgres"
	httpserver "github.com/eclesiaste/event-processor/internal/interfaces/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	db, err := postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	eventRepository := postgres.NewEventRepository(db)
	eventService := application.NewEventService(eventRepository)
	eventHandler := httpserver.NewEventHandler(eventService)

	server := httpserver.NewServer(cfg, eventHandler)

	log.Printf(
		"Event Processor running on port %s (%s)",
		cfg.HTTPPort,
		cfg.AppEnv,
	)

	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
