package recipes

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"

	"recipes-desk/internal/modules/recipes/handler"
	"recipes-desk/internal/modules/recipes/repository"
	"recipes-desk/internal/modules/recipes/service"
)

// Module represents the recipes module.
type Module struct {
	handler *handler.Handler
}

// NewModule creates a new recipes module.
func NewModule(db *mongo.Database, log *zerolog.Logger) *Module {
	recipeRepo := repository.NewMongoRepository(db)
	recipeService := service.NewService(recipeRepo, log)
	recipeHandler := handler.NewHandler(recipeService, log)

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
