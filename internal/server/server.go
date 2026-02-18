package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"recipes-desk/internal/config"
)

type server struct {
	log          *zerolog.Logger
	srv          *http.Server
	isProduction bool
}

func NewRouter() *gin.Engine {
	router := gin.Default()
	router.Use(cors.Default())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return router
}

func New(handler http.Handler, cfg *config.Config, logger *zerolog.Logger) *server {
	return &server{
		log: logger,
		srv: &http.Server{
			Addr:         ":" + cfg.Server.Port,
			Handler:      handler,
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
		},
		isProduction: cfg.IsProduction(),
	}
}

func (s *server) Start(ctx context.Context) {
	// Set Gin mode
	if s.isProduction {
		gin.SetMode(gin.ReleaseMode)
	}

	// Graceful shutdown
	go func() {
		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	s.log.Info().Str("address", s.srv.Addr).Msg("Server started")
}

func (s *server) Wait(ctx context.Context) {
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.log.Info().Msg("Shutting down server...")

	if err := s.srv.Shutdown(ctx); err != nil {
		s.log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	s.log.Info().Msg("Server exited")
}
