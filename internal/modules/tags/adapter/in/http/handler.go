package httphandler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"recipes-desk/internal/modules/tags/application"
	"recipes-desk/pkg/pagination"
)

// Handler handles HTTP requests for tags.
type Handler struct {
	service *application.Service
	log     *zerolog.Logger
}

// NewHandler creates a new tag handler.
func NewHandler(service *application.Service, log *zerolog.Logger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

// handleError maps application errors to HTTP status codes.
func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, application.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, application.ErrValidation):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, application.ErrConflict):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": application.ErrInternal.Error()})
	}
}

// Search handles GET /tags.
func (h *Handler) Search(c *gin.Context) {
	var req SearchTagsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.SearchTags(c.Request.Context(), req.Query, pagination.Request{
		Page:  req.Page,
		Limit: req.Limit,
	})
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toPaginatedResponse(result))
}

// GetByID handles GET /tags/:id.
func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")

	tag, err := h.service.GetTagByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toTagResponse(tag))
}

// Create handles POST /tags (for manual tag creation).
func (h *Handler) Create(c *gin.Context) {
	var req CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tag, err := h.service.CreateTag(c.Request.Context(), req.Name)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toTagResponse(tag))
}
