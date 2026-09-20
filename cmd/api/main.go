package main

import (
	"fmt"
	"log"

	"github.com/eclesiaste/event-processor/internal/config"
)

func main() {
	fmt.Println("Event processor starting...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Environment:", cfg.AppEnv)
	fmt.Println("HTTP Port:", cfg.HTTPPort)
}
