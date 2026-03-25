// Package out defines outbound interfaces for the recipes module.
package out

import (
	"context"

	"recipes-desk/internal/modules/recipes/domain/events"
)

// EventPublisher defines the interface for publishing recipe domain events.
type EventPublisher interface {
	// PublishRecipeCreated publishes a recipe.created event.
	PublishRecipeCreated(ctx context.Context, event events.RecipeCreated) error

	// PublishRecipeDeleted publishes a recipe.deleted event.
	PublishRecipeDeleted(ctx context.Context, event events.RecipeDeleted) error
}
