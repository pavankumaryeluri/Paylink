package main

import (
	"context"
	"fmt"
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
		util.Logger.Error("Failed to connect to database", "error", err)
		fmt.Printf("Database connection failed: %v\n", err)
		fmt.Println("Starting worker in demo mode...")
		database = &db.DB{
			Redis: &db.RedisClient{Addr: fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)},
		}
	}
	defer func() {
		if database != nil {
			database.Close()
		}
	}()

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
	fmt.Println("PayLink Worker running...")
	worker.Process(ctx)
}
