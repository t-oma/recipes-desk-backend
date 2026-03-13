package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"recipes-desk/internal/config"
)

type Server struct {
	log          *zerolog.Logger
	srv          *http.Server
	isProduction bool
}

func NewRouter() *gin.Engine {
	router := gin.Default()

	// CORS configuration with credentials support
	corsConfig := cors.Config{ //nolint:exhaustruct // using sensible defaults for other fields
		AllowOrigins:     []string{"http://localhost:5173"}, // Frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}
	router.Use(cors.New(corsConfig))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return router
}

func New(handler http.Handler, cfg *config.Config, logger *zerolog.Logger) *Server {
	return &Server{
		log: logger,
		srv: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
			Handler:      handler,
			ReadTimeout:  cfg.Server.Timeouts.Read,
			WriteTimeout: cfg.Server.Timeouts.Write,
		},
		isProduction: cfg.IsProduction(),
	}
}

func (s *Server) Start(ctx context.Context) {
	// Set Gin mode
	if s.isProduction {
		gin.SetMode(gin.ReleaseMode)
	}

	// Graceful shutdown
	go func() {
		err := s.srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	s.log.Info().Str("address", s.srv.Addr).Msg("Server started")

	s.waitSignal(ctx)
}

func (s *Server) waitSignal(ctx context.Context) {
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
