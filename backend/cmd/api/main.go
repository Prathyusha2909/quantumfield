package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Prathyusha2909/quantumfield/internal/auth"
	"github.com/Prathyusha2909/quantumfield/internal/config"
	"github.com/Prathyusha2909/quantumfield/internal/database"
	"github.com/Prathyusha2909/quantumfield/internal/httpapi"
	"github.com/Prathyusha2909/quantumfield/internal/queue"
)

func main() {
	cfg := config.Load()
	if len(cfg.JWTSecret) < 32 {
		log.Print("warning: JWT_SECRET should contain at least 32 characters")
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	if cfg.SeedDemo {
		database.SeedDemo(db)
	}

	queueClient := queue.New(cfg.RedisAddr, cfg.RedisPass, cfg.RedisDB)
	defer queueClient.Close()

	handler := &httpapi.Handler{
		DB:    db,
		Queue: queueClient,
		Auth:  auth.New(cfg.JWTSecret, cfg.JWTTTL),
	}
	router := httpapi.NewRouter(handler, cfg)

	address := ":" + cfg.APIPort
	log.Printf("QuantumField API listening on %s", address)
	if err := router.Run(address); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
