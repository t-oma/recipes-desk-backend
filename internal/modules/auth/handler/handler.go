package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"recipes-desk/internal/modules/auth/domain"
	"recipes-desk/internal/modules/auth/service"
)

type Handler struct {
	service AuthService
	log     *zerolog.Logger
}

type AuthService interface {
	GetByID(ctx context.Context, id string) (*service.SafeUser, error)
	GetByEmail(ctx context.Context, email string) (*service.SafeUser, error)
	Register(ctx context.Context, params *service.RegisterParams) (*service.AuthResult, error)
	Login(ctx context.Context, params *service.LoginParams) (*service.AuthResult, error)
}

func NewHandler(service AuthService, log *zerolog.Logger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
	case errors.Is(err, domain.ErrAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
	case errors.Is(err, domain.ErrValidation):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Debug().Err(err).Msg("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.Register(c.Request.Context(), &service.RegisterParams{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Password:  req.Password,
	})
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toResponse(result))
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Debug().Err(err).Msg("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.log.Debug().Str("user_email", req.Email).Msg("Logging in user")
	result, err := h.service.Login(c.Request.Context(), &service.LoginParams{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(result))
}
