package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/vibeswithkk/paylink/internal/api"
	"github.com/vibeswithkk/paylink/internal/config"
	"github.com/vibeswithkk/paylink/internal/db"
	"github.com/vibeswithkk/paylink/internal/util"
)

func main() {
	util.InitLogger()
	util.Logger.Info("Starting PayLink Server...")

	cfg := config.LoadConfig()

	database, err := db.Connect(cfg)
	if err != nil {
		util.Logger.Error("Failed to connect to database", "error", err)
		fmt.Printf("Database connection failed: %v\n", err)
		fmt.Println("Starting in demo mode without database...")
		database = &db.DB{}
	}
	defer func() {
		if database != nil {
			database.Close()
		}
	}()

	handler := api.NewHandler(cfg, database)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler.Routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	util.Logger.Info("Server listening", "port", cfg.Port)
	fmt.Printf("PayLink API Server running at http://localhost:%s\n", cfg.Port)

	if err := srv.ListenAndServe(); err != nil {
		util.Logger.Error("Server failed", "error", err)
	}
}
