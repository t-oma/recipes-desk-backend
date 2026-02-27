package auth

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"

	"recipes-desk/internal/modules/auth/handler"
	"recipes-desk/internal/modules/auth/repository/mongorepo"
	"recipes-desk/internal/modules/auth/service"
)

// Module represents the authentication module.
type Module struct {
	handler    *handler.Handler
	middleware gin.HandlerFunc
}

// NewModule creates a new auth module.
func NewModule(
	db *mongo.Database,
	log *zerolog.Logger,
	secret string,
	accessExpiry time.Duration,
	refreshExpiry time.Duration,
) *Module {
	authRepo := mongorepo.NewUsers(db)
	refreshRepo := mongorepo.NewRefreshTokens(db)

	passwordService := service.NewBcryptHasher(14)
	tokenService := service.NewTokenService(refreshRepo, secret, accessExpiry, refreshExpiry)
	authService := service.NewService(authRepo, log, passwordService, tokenService)

	authHandler := handler.NewHandler(authService, log)

	return &Module{
		handler:    authHandler,
		middleware: handler.AuthMiddleware(tokenService),
	}
}

// Middleware returns the auth middleware.
func (m *Module) Middleware() gin.HandlerFunc {
	return m.middleware
}

// RegisterRoutes registers all auth routes.
func (m *Module) RegisterRoutes(public, protected *gin.RouterGroup) {
	// Public routes
	public.POST("/auth/register", m.handler.Register)
	public.POST("/auth/login", m.handler.Login)
	public.POST("/auth/refresh", m.handler.Refresh)

	// Protected routes (require authentication)
	protected.POST("/auth/logout", m.handler.Logout)
	protected.GET("/auth/me", m.handler.Me)
}
