package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog"

	"recipes-desk/internal/infra/messagebus"
	"recipes-desk/internal/modules/tags/application/ports/in"
	"recipes-desk/internal/modules/tags/domain/events"
)

type EventHandler struct {
	service in.TagService
	log     *zerolog.Logger
}

// NewEventHandler creates a new recipe event consumer.
func NewEventHandler(
	service in.TagService,
	log *zerolog.Logger,
) *EventHandler {
	return &EventHandler{
		service: service,
		log:     log,
	}
}

// HandleRecipeCreated processes recipe.created events.
func (c *EventHandler) HandleRecipeCreated(msg messagebus.Message) error {
	var event events.RecipeCreated
	if err := json.Unmarshal(msg.Payload, &event); err != nil {
		c.log.Error().Err(err).Msg("Failed to unmarshal recipe created event")
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	if err := c.service.EnsureTagsExist(context.Background(), event.Tags); err != nil {
		c.log.Error().
			Err(err).
			Str("recipe_id", event.RecipeID).
			Msg("Failed to ensure tags exist")
		return err
	}

	c.log.Info().
		Str("recipe_id", event.RecipeID).
		Strs("tags", event.Tags).
		Msg("Successfully processed recipe created event")

	return nil
}

// HandleRecipeDeleted processes recipe.deleted events.
func (c *EventHandler) HandleRecipeDeleted(msg messagebus.Message) error {
	var event events.RecipeDeleted
	if err := json.Unmarshal(msg.Payload, &event); err != nil {
		c.log.Error().Err(err).Msg("Failed to unmarshal recipe deleted event")
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	// For now, we don't delete tags even if they have no recipes.
	// In the future, we can implement usage count and delete orphaned tags.
	c.log.Info().
		Str("recipe_id", event.RecipeID).
		Msg("Recipe deleted - tags usage count not updated yet")

	return nil
}
