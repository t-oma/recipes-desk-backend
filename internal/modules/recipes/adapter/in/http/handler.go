package httphandler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"recipes-desk/internal/modules/recipes/application/dto"
	"recipes-desk/internal/modules/recipes/application/ports/in"
	"recipes-desk/internal/modules/recipes/domain"
)

// Handler handles HTTP requests for recipes.
type Handler struct {
	service in.RecipeService
	log     *zerolog.Logger
}

// NewHandler creates a new recipe handler.
func NewHandler(service in.RecipeService, log *zerolog.Logger) *Handler {
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

func (h *Handler) List(c *gin.Context) {
	h.log.Debug().Msg("Getting all recipes")

	result, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to get recipes")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	response := make([]RecipeResponse, len(result))
	for i, recipe := range result {
		response[i] = *toRecipeResponse(&recipe)
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) Search(c *gin.Context) {
	query := c.Query("q")
	h.log.Debug().Str("query", query).Msg("Searching recipes")

	result, err := h.service.Search(c.Request.Context(), query)
	if err != nil {
		h.log.Error().Err(err).Str("query", query).Msg("Failed to search recipes")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	response := make([]RecipeResponse, len(result))
	for i, recipe := range result {
		response[i] = *toRecipeResponse(&recipe)
	}

	c.JSON(http.StatusOK, response)
}

// GetByID handles GET /recipes/:id.
func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")
	h.log.Debug().Str("recipe_id", id).Msg("Getting recipe")

	result, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toRecipeResponse(result))
}

func (h *Handler) Create(c *gin.Context) {
	userID := c.GetString("userID")
	_ = userID

	var req CreateRecipeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Debug().Err(err).Msg("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.Create(c.Request.Context(), dto.CreateRecipeInput{
		Title:       req.Title,
		Description: req.Description,
		Ingredients: mapIngredients(req.Ingredients),
		Steps:       mapSteps(req.Steps),
		CookingTime: req.CookingTime,
		Portions:    req.Portions,
		Tags:        req.Tags,
	})
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toRecipeResponse(result))
}

func (h *Handler) Update(c *gin.Context) {
	userID := c.GetString("userID")
	_ = userID

	id := c.Param("id")

	var req CreateRecipeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Debug().Err(err).Msg("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.log.Debug().Str("recipe_id", id).Msg("Updating recipe")

	result, err := h.service.Update(c.Request.Context(), id, dto.UpdateRecipeInput{
		Title:       req.Title,
		Description: req.Description,
		Ingredients: mapIngredients(req.Ingredients),
		Steps:       mapSteps(req.Steps),
		CookingTime: req.CookingTime,
		Portions:    req.Portions,
		Tags:        req.Tags,
	})
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toRecipeResponse(result))
}

func (h *Handler) Delete(c *gin.Context) {
	userID := c.GetString("userID")
	_ = userID

	id := c.Param("id")
	h.log.Debug().Str("recipe_id", id).Msg("Deleting recipe")

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "recipe deleted successfully"})
}
