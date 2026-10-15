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

	"github.com/joho/godotenv"

	"github.com/eclesiaste/event-processor/internal/application"
	"github.com/eclesiaste/event-processor/internal/config"
	"github.com/eclesiaste/event-processor/internal/infrastructure/kafka"
	"github.com/eclesiaste/event-processor/internal/infrastructure/migrations"
	"github.com/eclesiaste/event-processor/internal/infrastructure/postgres"
	httpserver "github.com/eclesiaste/event-processor/internal/interfaces/http"
)

func main() {
	godotenv.Load()

	configurations, error := config.Load()
	if error != nil {
		log.Fatal(error)
	}

	appContext, stop := context.WithCancel(context.Background())
	defer stop()

	database, error := postgres.New(appContext, configurations.DatabaseURL)
	if error != nil {
		log.Fatal(error)
	}
	defer database.Close()

	if error := migrations.Run(configurations.DatabaseURL); error != nil {
		log.Fatal(error)
	}

	producer, error := kafka.NewProducer(
		configurations.KafkaBrokers,
		configurations.KafkaTopic,
	)
	if error != nil {
		log.Fatal(error)
	}
	defer func() {
		if error := producer.Close(); error != nil {
			log.Printf("Kafka producer close error: %v", error)
		}
	}()

	consumer, error := kafka.NewConsumer(
		configurations.KafkaBrokers,
		configurations.KafkaTopic,
		configurations.KafkaGroupID,
	)
	if error != nil {
		log.Fatal(error)
	}
	defer func() {
		if error := consumer.Close(); error != nil {
			log.Printf("Kafka consumer close error: %v", error)
		}
	}()

	eventRepository := postgres.NewEventRepository(database)
	eventService := application.NewEventService(eventRepository, producer)
	eventHandler := httpserver.NewEventHandler(eventService)

	server := httpserver.NewServer(configurations, eventHandler)

	log.Printf(
		"Event Processor running on port %s (%s)",
		configurations.HTTPPort,
		configurations.AppEnv,
	)

	consumerStopped := make(chan struct{})

	go func() {
		defer close(consumerStopped)

		if error := consumer.Start(appContext); error != nil {
			log.Printf("Kafka consumer error: %v", error)
		}
	}()

	go func() {
		if error := server.Start(); error != nil && !errors.Is(error, http.ErrServerClosed) {
			log.Printf("HTTP server error: %v", error)
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

	stop()

	shutdownContext, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	select {
	case <-consumerStopped:
	case <-shutdownContext.Done():
		log.Printf("Kafka consumer shutdown timed out")
	}

	if error := server.Shutdown(shutdownContext); error != nil {
		log.Printf("HTTP server shutdown error: %v", error)
	}

	log.Println("Event Processor stopped")
}
