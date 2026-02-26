package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"recipes-desk/internal/modules/recipes/domain"
)

// Handler handles HTTP requests for recipes.
type Handler struct {
	service RecipeService
	log     *zerolog.Logger
}

// RecipeService defines the service interface.
type RecipeService interface {
	Create(ctx context.Context, recipe *domain.Recipe) (*domain.Recipe, error)
	GetByID(ctx context.Context, id string) (*domain.Recipe, error)
	GetAll(ctx context.Context) ([]domain.Recipe, error)
	Search(ctx context.Context, query string) ([]domain.Recipe, error)
	Update(ctx context.Context, id string, recipe *domain.Recipe) (*domain.Recipe, error)
	Delete(ctx context.Context, id string) error
}

// NewHandler creates a new recipe handler.
func NewHandler(service RecipeService, log *zerolog.Logger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

// handleError maps domain errors to HTTP status codes.
func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "recipe not found"})
	case errors.Is(err, domain.ErrValidation):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

// List handles GET /recipes.
func (h *Handler) List(c *gin.Context) {
	h.log.Debug().Msg("Getting all recipes")

	recipes, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to get recipes")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, toListResponse(recipes))
}

// Search handles GET /recipes/search.
func (h *Handler) Search(c *gin.Context) {
	query := c.Query("q")
	h.log.Debug().Str("query", query).Msg("Searching recipes")

	recipes, err := h.service.Search(c.Request.Context(), query)
	if err != nil {
		h.log.Error().Err(err).Str("query", query).Msg("Failed to search recipes")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, toListResponse(recipes))
}

// GetByID handles GET /recipes/:id.
func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")
	h.log.Debug().Str("recipe_id", id).Msg("Getting recipe")

	recipe, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(recipe))
}

// Create handles POST /recipes.
func (h *Handler) Create(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Debug().Err(err).Msg("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.log.Debug().Str("title", req.Title).Msg("Creating recipe")

	recipe := toDomainRecipe(req)
	created, err := h.service.Create(c.Request.Context(), recipe)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toResponse(created))
}

// Update handles PUT /recipes/:id.
func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")

	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Debug().Err(err).Msg("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.log.Debug().Str("recipe_id", id).Msg("Updating recipe")

	recipe := toDomainRecipe(req)
	updated, err := h.service.Update(c.Request.Context(), id, recipe)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(updated))
}

// Delete handles DELETE /recipes/:id.
func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	h.log.Debug().Str("recipe_id", id).Msg("Deleting recipe")

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "recipe deleted successfully"})
}
