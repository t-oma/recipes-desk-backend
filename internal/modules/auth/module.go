package auth

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"

	httphandler "recipes-desk/internal/modules/auth/adapter/in/http"
	"recipes-desk/internal/modules/auth/adapter/out/mongorepo"
	"recipes-desk/internal/modules/auth/application"
)

// Module represents the authentication module.
type Module struct {
	handler    *httphandler.Handler
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
	idGen := mongorepo.ObjectIDGenerator{}
	tokenService := application.NewTokenService(
		refreshRepo,
		idGen,
		log,
		secret,
		accessExpiry,
		refreshExpiry,
	)
	authService := application.NewService(authRepo, log, passwordService, tokenService, idGen)

	authHandler := httphandler.NewHandler(authService, log)

	return &Module{
		handler:    authHandler,
		middleware: httphandler.AuthMiddleware(tokenService),
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
