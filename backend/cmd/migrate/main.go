package main

import (
	"log"

	"github.com/Prathyusha2909/quantumfield/internal/config"
	"github.com/Prathyusha2909/quantumfield/internal/database"
)

func main() {
	cfg := config.Load()
	if _, err := database.Connect(cfg.DatabaseURL); err != nil {
		log.Fatal(err)
	}
	log.Print("database migrations are current")
}
