package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/Prathyusha2909/quantumfield/internal/config"
	"github.com/Prathyusha2909/quantumfield/internal/database"
	"github.com/Prathyusha2909/quantumfield/internal/queue"
	"github.com/Prathyusha2909/quantumfield/internal/scanner"
	"github.com/Prathyusha2909/quantumfield/internal/worker"
)

func main() {
	cfg := config.Load()
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	queueClient := queue.New(cfg.RedisAddr, cfg.RedisPass, cfg.RedisDB)
	defer queueClient.Close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	service := &worker.Worker{
		DB:      db,
		Queue:   queueClient,
		Scanner: scanner.New(cfg.ScanTimeout),
	}
	if err := service.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
