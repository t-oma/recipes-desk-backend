package tags

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"

	httphandler "recipes-desk/internal/modules/tags/adapter/in/http"
	"recipes-desk/internal/modules/tags/adapter/out/messaging"
	"recipes-desk/internal/modules/tags/adapter/out/mongorepo"
	"recipes-desk/internal/modules/tags/application"
	"recipes-desk/internal/modules/tags/config"
)

// Module represents the tags module.
type Module struct {
	handler       *httphandler.Handler
	eventConsumer *messaging.RecipeEventConsumer
}

// NewModule creates a new tags module.
func NewModule(db *mongo.Database, log *zerolog.Logger) *Module {
	config, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load tags config")
	}

	tagRepo := mongorepo.NewTags(db)
	err = tagRepo.InitIndexes(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize tag indexes")
	}
	idGenerator := mongorepo.ObjectIDGenerator{}
	tagService := application.NewService(
		tagRepo,
		idGenerator,
		log,
		config.Pagination.MaxLimit,
		config.Pagination.DefaultLimit,
	)
	tagHandler := httphandler.NewHandler(tagService, log)

	eventConsumer := messaging.NewRecipeEventConsumer(tagService, log)

	return &Module{
		handler:       tagHandler,
		eventConsumer: eventConsumer,
	}
}

// RegisterRoutes registers all tag routes.
func (m *Module) RegisterRoutes(public, protected *gin.RouterGroup) {
	// Public routes
	public.GET("/tags", m.handler.Search)
	public.GET("/tags/:id", m.handler.GetByID)

	// Protected routes (admin only in future)
	protected.POST("/tags", m.handler.Create)
}

// GetEventConsumer returns the event consumer for registration with message bus.
func (m *Module) GetEventConsumer() *messaging.RecipeEventConsumer {
	return m.eventConsumer
}
