package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"recipes-desk/internal/config"
	"recipes-desk/internal/infra/database"
	"recipes-desk/internal/infra/logger"
	"recipes-desk/internal/server"
)

// @title Recipes Desk API
// @version 1.0
// @description API for managing recipes and user accounts
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@recipes-desk.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.New(cfg.App.Environment)
	log.Debug().Any("config", cfg).Msg("Loaded configuration")

	// Connect to MongoDB
	db, err := database.New(&cfg.MongoDB, log)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}
	defer func() {
		if err = db.Disconnect(ctx); err != nil {
			log.Fatal().Err(err).Msg("Failed to disconnect from MongoDB")
		}
	}()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	server.New(server.NewRouter(), cfg, log).Start(ctx)
}
