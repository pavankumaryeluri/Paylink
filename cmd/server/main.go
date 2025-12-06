package main

import (
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
		util.Logger.Error("Failed to connect to database/redis", "error", err)
		panic(err)
	}
	defer database.Close()

	handler := api.NewHandler(cfg, database)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler.Routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	util.Logger.Info("Server listening", "port", cfg.Port)
	if err := srv.ListenAndServe(); err != nil {
		util.Logger.Error("Server failed", "error", err)
	}
}
