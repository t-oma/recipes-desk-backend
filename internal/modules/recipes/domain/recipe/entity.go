// Package recipe contains recipe entity, repository, value objects and domain errors.
package recipe

import (
	"errors"
	"fmt"
	"time"
)

// Entity represents a cooking recipe.
type Entity struct { //nolint:recvcheck // intentionally mixed pointer and value receivers
	id          EntityID
	title       Title
	description Description
	ingredients []Ingredient
	steps       []Step
	cookingTime CookingTime
	portions    Portions
	tags        []Tag
	authorID    AuthorID
	createdAt   time.Time
	updatedAt   time.Time
}

func NewEntity(
	id string,
	title string,
	description string,
	ingredientsVO []Ingredient,
	stepsVO []Step,
	cookingTime int64,
	portions int,
	tagsVO []Tag,
	authorID string,
) (*Entity, error) {
	if len(ingredientsVO) == 0 {
		return nil, ErrNoIngredients
	}
	if len(stepsVO) == 0 {
		return nil, ErrNoSteps
	}
	if len(tagsVO) == 0 {
		return nil, ErrNoTags
	}

	idVO, err := NewEntityID(id)
	if err != nil {
		return nil, err
	}

	titleVO, err := NewTitle(title)
	if err != nil {
		return nil, err
	}

	descVO, err := NewDescription(description)
	if err != nil {
		return nil, err
	}

	cookingTimeVO, err := NewCookingTime(cookingTime)
	if err != nil {
		return nil, err
	}

	portionsVO, err := NewPortions(portions)
	if err != nil {
		return nil, err
	}

	authorIDVO, err := NewAuthorID(authorID)
	if err != nil {
		return nil, err
	}

	return &Entity{
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

func (r *Entity) AddIngredient(ingredient Ingredient) error {
	for _, existing := range r.ingredients {
		if existing.Name() == ingredient.Name() {
			return ErrDuplicateIngredient
		}
	}
	r.ingredients = append(r.ingredients, ingredient)
	return nil
}

func (r *Entity) AddStep(step Step) error {
	expectedOrder := len(r.steps) + 1
	if step.Order() != expectedOrder {
		return ErrInvalidStepOrderf(expectedOrder, step.Order())
	}
	r.steps = append(r.steps, step)
	return nil
}

func (r *Entity) AddTag(tag Tag) {
	r.tags = append(r.tags, tag)
}

func (r Entity) ID() EntityID {
	return r.id
}

func (r Entity) AuthorID() AuthorID {
	return r.authorID
}

func (r Entity) Title() Title {
	return r.title
}

func (r Entity) Description() Description {
	return r.description
}

func (r Entity) Ingredients() []Ingredient {
	copied := make([]Ingredient, len(r.ingredients))
	copy(copied, r.ingredients)
	return copied
}

func (r Entity) Steps() []Step {
	copied := make([]Step, len(r.steps))
	copy(copied, r.steps)
	return copied
}

func (r Entity) CookingTime() CookingTime {
	return r.cookingTime
}

func (r Entity) Portions() Portions {
	return r.portions
}

func (r Entity) Tags() []Tag {
	copied := make([]Tag, len(r.tags))
	copy(copied, r.tags)
	return copied
}

func (r Entity) UpdatedAt() time.Time {
	return r.updatedAt
}

func (r Entity) CreatedAt() time.Time {
	return r.createdAt
}

func (r Entity) HasID() bool {
	return r.ID().Value() != ""
}

func (r *Entity) AssignID(id EntityID) {
	r.id = id
}

func (r Entity) Equals(other *Entity) bool {
	if other == nil {
		return false
	}

	return r.ID().Value() == other.ID().Value()
}

func (r *Entity) RestoreFromPersistence(updatedAt time.Time, createdAt time.Time) {
	r.updatedAt = updatedAt
	r.createdAt = createdAt
}

var ErrValidation = errors.New("validation error")

// Validation errors - wrapped with ErrValidation category.
var (
	ErrNoIngredients = fmt.Errorf(
		"%w: recipe must have at least one ingredient",
		ErrValidation,
	)
	ErrNoSteps = fmt.Errorf(
		"%w: recipe must have at least one step",
		ErrValidation,
	)
	ErrNoTags              = fmt.Errorf("%w: recipe must have at least one tag", ErrValidation)
	ErrDuplicateIngredient = fmt.Errorf(
		"%w: recipe cannot have duplicate ingredients",
		ErrValidation,
	)
)

func ErrInvalidStepOrderf(expectedOrder int, actualOrder int) error {
	return fmt.Errorf(
		"%w: step order must be %d got %d",
		ErrValidation,
		expectedOrder,
		actualOrder,
	)
}
