package auth

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"

	"recipes-desk/internal/modules/auth/application"
	"recipes-desk/internal/modules/auth/handler"
	"recipes-desk/internal/modules/auth/repository/mongorepo"
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	authRepo := mongorepo.NewUsers(db)
	refreshRepo := mongorepo.NewRefreshTokens(db)
	err := refreshRepo.InitIndexes(ctx)
	if err != nil {
		cancel()
		log.Fatal(). //nolint:gocritic // cancel() is called
				Err(err).
				Msg("Failed to initialize refresh token indexes")
	}

	passwordService := application.NewBcryptHasher(14)
	tokenService := application.NewTokenService(refreshRepo, secret, accessExpiry, refreshExpiry)
	authService := application.NewService(authRepo, log, passwordService, tokenService)

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
