package recipes

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"

	httphandler "recipes-desk/internal/modules/recipes/adapter/in/http"
	"recipes-desk/internal/modules/recipes/adapter/out/mongorepo"
	"recipes-desk/internal/modules/recipes/application"
	"recipes-desk/internal/modules/recipes/config"
)

// Module represents the recipes module.
type Module struct {
	handler *httphandler.Handler
}

// NewModule creates a new recipes module.
func NewModule(db *mongo.Database, log *zerolog.Logger) *Module {
	config, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load recipes config")
	}

	recipeRepo := mongorepo.NewRecipes(db)
	idGenerator := mongorepo.ObjectIDGenerator{}
	recipeService := application.NewService(
		recipeRepo,
		idGenerator,
		log,
		config.Pagination.MaxLimit,
		config.Pagination.DefaultLimit,
	)
	recipeHandler := httphandler.NewHandler(recipeService, log)

	return &Module{
		handler: recipeHandler,
	}
}

// RegisterRoutes registers all recipe routes.
func (m *Module) RegisterRoutes(public, protected *gin.RouterGroup) {
	// Public routes
	public.GET("/recipes", m.handler.List)
	public.GET("/recipes/search", m.handler.Search)
	public.GET("/recipes/:id", m.handler.GetByID)

	// Protected routes (will require auth middleware later)
	protected.POST("/recipes", m.handler.Create)
	protected.PUT("/recipes/:id", m.handler.Update)
	protected.DELETE("/recipes/:id", m.handler.Delete)
}
