package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"recipes-desk/internal/config"
	"recipes-desk/internal/infra/database"
	"recipes-desk/internal/infra/logger"
	"recipes-desk/internal/infra/messagebus/rabbitmq"
	"recipes-desk/internal/modules/auth"
	"recipes-desk/internal/modules/recipes"
	"recipes-desk/internal/modules/tags"
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
			panic(err)
		}
	}()

	rmqConn, err := rabbitmq.NewConnection(cfg.RabbitMQ, log)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = rmqConn.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close RabbitMQ connection")
		}
	}()

	// Setup router and register modules
	router := server.NewRouter()
	api := router.Group("/api/v1")

	// Create public and protected route groups
	public := api.Group("")
	protected := api.Group("")

	// Initialize and register auth module
	authModule := auth.NewModule(
		db.Database,
		log,
	)
	protected.Use(authModule.Middleware())
	authModule.RegisterRoutes(public, protected)

	// Initialize and register recipes module
	recipesModule := recipes.NewModule(db.Database, rmqConn, log)
	recipesModule.RegisterRoutes(public, protected)

	// Initialize and register tags module
	tagsModule := tags.NewModule(db.Database, rmqConn, log)
	tagsModule.RegisterRoutes(public, protected)

	defer func() {
		if err = recipesModule.Shutdown(); err != nil {
			log.Error().Err(err).Msg("Failed to shutdown recipes module")
		}
		if err = tagsModule.Shutdown(); err != nil {
			log.Error().Err(err).Msg("Failed to shutdown tags module")
		}
	}()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	server.New(router, cfg, log).Start(ctx)
}
