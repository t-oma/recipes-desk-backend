// Package entity contains recipe entity, repository, value objects and domain errors.
package entity

import (
	"fmt"
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
	id string,
	title string,
	description string,
	ingredientsVO []vo.Ingredient,
	stepsVO []vo.Step,
	cookingTime int64,
	portions int,
	tagsVO []vo.Tag,
	authorID string,
) (*Recipe, error) {
	if len(ingredientsVO) == 0 {
		return nil, ErrNoIngredients
	}
	if len(stepsVO) == 0 {
		return nil, ErrNoSteps
	}
	if len(tagsVO) == 0 {
		return nil, ErrNoTags
	}

	idVO, err := vo.NewEntityID(id)
	if err != nil {
		return nil, err
	}

	titleVO, err := vo.NewTitle(title)
	if err != nil {
		return nil, err
	}

	descVO, err := vo.NewDescription(description)
	if err != nil {
		return nil, err
	}

	cookingTimeVO, err := vo.NewCookingTime(cookingTime)
	if err != nil {
		return nil, err
	}

	portionsVO, err := vo.NewPortions(portions)
	if err != nil {
		return nil, err
	}

	authorIDVO, err := vo.NewAuthorID(authorID)
	if err != nil {
		return nil, err
	}

	return &Recipe{
		id:          idVO,
		title:       titleVO,
		description: descVO,
		ingredients: ingredientsVO,
		steps:       stepsVO,
		cookingTime: cookingTimeVO,
		portions:    portionsVO,
		tags:        tagsVO,
		authorID:    authorIDVO,
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
			return ErrDuplicateIngredient
		}
	}
	r.ingredients = append(r.ingredients, ingredient)
	return nil
}

func (r *Recipe) AddStep(step vo.Step) error {
	expectedOrder := len(r.steps) + 1
	if step.Order() != expectedOrder {
		return ErrInvalidStepOrderf(expectedOrder, step.Order())
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

// Validation errors - wrapped with ErrValidation category.
var (
	ErrNoIngredients = fmt.Errorf(
		"%w: recipe must have at least one ingredient",
		domain.ErrValidation,
	)
	ErrNoSteps = fmt.Errorf(
		"%w: recipe must have at least one step",
		domain.ErrValidation,
	)
	ErrNoTags = fmt.Errorf(
		"%w: recipe must have at least one tag",
		domain.ErrValidation,
	)
	ErrDuplicateIngredient = fmt.Errorf(
		"%w: recipe cannot have duplicate ingredients",
		domain.ErrValidation,
	)
)

func ErrInvalidStepOrderf(expectedOrder int, actualOrder int) error {
	return fmt.Errorf(
		"%w: step order must be %d got %d",
		domain.ErrValidation,
		expectedOrder,
		actualOrder,
	)
}
