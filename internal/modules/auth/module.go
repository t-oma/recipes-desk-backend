package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"

	"recipes-desk/internal/modules/auth/handler"
	"recipes-desk/internal/modules/auth/repository"
	"recipes-desk/internal/modules/auth/service"
)

type Module struct {
	handler *handler.Handler
}

func NewModule(db *mongo.Database, log *zerolog.Logger) *Module {
	authRepo := repository.NewMongoRepository(db)
	passwordService := service.NewBcryptHasher(14)
	authService := service.NewService(authRepo, log, passwordService)
	authHandler := handler.NewHandler(authService, log)

	return &Module{
		handler: authHandler,
	}
}

// RegisterRoutes registers all auth routes.
func (m *Module) RegisterRoutes(public, _ *gin.RouterGroup) {
	public.POST("/auth/register", m.handler.Register)
	public.POST("/auth/login", m.handler.Login)
}
