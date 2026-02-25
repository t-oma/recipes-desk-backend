package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"

	"recipes-desk/internal/config"
	"recipes-desk/internal/modules/auth/handler"
	"recipes-desk/internal/modules/auth/repository"
	"recipes-desk/internal/modules/auth/service"
)

type Module struct {
	handler    *handler.Handler
	middleware gin.HandlerFunc
}

func NewModule(db *mongo.Database, log *zerolog.Logger, cfg *config.JWT) *Module {
	passwordService := service.NewBcryptHasher(14)
	tokenService := service.NewTokenService(cfg.Secret, cfg.AccessExpiry)

	authRepo := repository.NewMongoRepository(db)
	authService := service.NewService(authRepo, log, passwordService, tokenService)
	authHandler := handler.NewHandler(authService, log)

	return &Module{
		handler:    authHandler,
		middleware: handler.AuthMiddleware(tokenService),
	}
}

func (m *Module) Middleware() gin.HandlerFunc {
	return m.middleware
}

// RegisterRoutes registers all auth routes.
func (m *Module) RegisterRoutes(public, protected *gin.RouterGroup) {
	// Public routes
	public.POST("/auth/register", m.handler.Register)
	public.POST("/auth/login", m.handler.Login)

	// Protected routes (require authentication)
	protected.POST("/auth/logout", m.handler.Logout)
	protected.GET("/auth/me", m.handler.Me)
}
