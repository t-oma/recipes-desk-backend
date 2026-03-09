// Package entity contains recipe entity definitions.
package entity

import (
	"time"

	"recipes-desk/internal/modules/recipes/domain"
	vo "recipes-desk/internal/modules/recipes/domain/valueobject"
)

// Recipe represents a cooking recipe.
type Recipe struct { //nolint:recvcheck // intentionally mixed pointer and value receivers
	id          vo.EntityID
	title       vo.Title
	description vo.Description
	ingredients []vo.Ingredient
	steps       []vo.Step
	cookingTime vo.CookingTime
	portions    vo.Portions
	tags        []vo.Tag
	authorID    vo.AuthorID
	createdAt   time.Time
	updatedAt   time.Time
}

func NewRecipe(
	id vo.EntityID,
	title vo.Title,
	description vo.Description,
	ingredients []vo.Ingredient,
	steps []vo.Step,
	cookingTime vo.CookingTime,
	portions vo.Portions,
	tags []vo.Tag,
	authorID vo.AuthorID,
) (*Recipe, error) {
	if len(ingredients) == 0 {
		return nil, domain.ErrNoIngredients
	}
	if len(steps) == 0 {
		return nil, domain.ErrNoSteps
	}
	if len(tags) == 0 {
		return nil, domain.ErrNoTags
	}

	return &Recipe{
		id:          id,
		title:       title,
		description: description,
		ingredients: ingredients,
		steps:       steps,
		cookingTime: cookingTime,
		portions:    portions,
		tags:        tags,
		authorID:    authorID,
		createdAt:   time.Time{},
		updatedAt:   time.Time{},
	}, nil
}

func (r *Recipe) UpdateTitle(title vo.Title) {
	r.title = title
}

func (r *Recipe) UpdateDescription(description vo.Description) {
	r.description = description
}

func (r *Recipe) AddIngredient(ingredient vo.Ingredient) error {
	for _, existing := range r.ingredients {
		if existing.Name() == ingredient.Name() {
			return domain.ErrDuplicateIngredient
		}
	}
	r.ingredients = append(r.ingredients, ingredient)
	return nil
}

func (r *Recipe) AddStep(step vo.Step) error {
	expectedOrder := len(r.steps) + 1
	if step.Order() != expectedOrder {
		return domain.ErrInvalidStepOrderf(expectedOrder, step.Order())
	}
	r.steps = append(r.steps, step)
	return nil
}

func (r *Recipe) AddTag(tag vo.Tag) {
	r.tags = append(r.tags, tag)
}

func (r Recipe) ID() vo.EntityID {
	return r.id
}

func (r Recipe) AuthorID() vo.AuthorID {
	return r.authorID
}

func (r Recipe) Title() vo.Title {
	return r.title
}

func (r Recipe) Description() vo.Description {
	return r.description
}

func (r Recipe) Ingredients() []vo.Ingredient {
	copied := make([]vo.Ingredient, len(r.ingredients))
	copy(copied, r.ingredients)
	return copied
}

func (r Recipe) Steps() []vo.Step {
	copied := make([]vo.Step, len(r.steps))
	copy(copied, r.steps)
	return copied
}

func (r Recipe) CookingTime() vo.CookingTime {
	return r.cookingTime
}

func (r Recipe) Portions() vo.Portions {
	return r.portions
}

func (r Recipe) Tags() []vo.Tag {
	copied := make([]vo.Tag, len(r.tags))
	copy(copied, r.tags)
	return copied
}

func (r Recipe) UpdatedAt() time.Time {
	return r.updatedAt
}

func (r Recipe) CreatedAt() time.Time {
	return r.createdAt
}

func (r Recipe) HasID() bool {
	return r.ID().String() != ""
}

func (r Recipe) CanBeModified(userID string) bool {
	return r.authorID.String() == userID
}

func (r *Recipe) AssignID(id vo.EntityID) {
	r.id = id
}

func (r Recipe) Equals(other *Recipe) bool {
	if other == nil {
		return false
	}

	return r.ID().String() == other.ID().String()
}

func (r *Recipe) RestoreFromPersistence(updatedAt time.Time, createdAt time.Time) {
	r.updatedAt = updatedAt
	r.createdAt = createdAt
}
