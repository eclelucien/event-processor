package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	go func() {
		if err := server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	signalChannel := make(chan os.Signal, 1)

	signal.Notify(
		signalChannel,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	<-signalChannel

	log.Println("Shutdown signal received")

	shutdownContext, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownContext); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	log.Println("Event Processor stopped")
}
