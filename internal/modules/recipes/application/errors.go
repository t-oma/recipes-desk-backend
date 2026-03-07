package application

import "recipes-desk/internal/modules/recipes/domain"

// Re-export domain errors for use by handlers.
// This provides a clean public API while keeping domain details encapsulated.
var (
	// ErrValidation indicates a validation error.
	ErrValidation = domain.ErrValidation
	// ErrNotFound indicates a resource was not found.
	ErrNotFound = domain.ErrNotFound
	// ErrForbidden indicates the operation is forbidden.
	ErrForbidden = domain.ErrForbidden
	// ErrRecipeNotFound indicates a recipe was not found.
	ErrRecipeNotFound = domain.ErrRecipeNotFound
	// ErrInternal indicates an internal server error.
	ErrInternal = domain.ErrInternal
)
