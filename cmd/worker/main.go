package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/vibeswithkk/paylink/internal/config"
	"github.com/vibeswithkk/paylink/internal/db"
	"github.com/vibeswithkk/paylink/internal/jobs"
	"github.com/vibeswithkk/paylink/internal/util"
)

func main() {
	util.InitLogger()
	util.Logger.Info("Starting PayLink Worker...")

	cfg := config.LoadConfig()

	database, err := db.Connect(cfg)
	if err != nil {
		util.Logger.Error("Failed to connect to database/redis", "error", err)
		panic(err)
	}
	defer database.Close()

	worker := jobs.NewWorker(database.Redis)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		util.Logger.Info("Shutting down worker...")
		cancel()
	}()

	util.Logger.Info("Worker processing jobs...")
	worker.Process(ctx)
}
