package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Recipe represents a cooking recipe.
type Recipe struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"` //nolint:tagliatelle // mongoDB id
	Title       string             `bson:"title"         json:"title"`
	Description string             `bson:"description"   json:"description"`
	Ingredients []Ingredient       `bson:"ingredients"   json:"ingredients"`
	Steps       []Step             `bson:"steps"         json:"steps"`
	CookingTime int                `bson:"cookingTime"   json:"cookingTime"`
	Portions    int                `bson:"portions"      json:"portions"`
	Tags        []string           `bson:"tags"          json:"tags"`
	CreatedAt   time.Time          `bson:"createdAt"     json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt"     json:"updatedAt"`
}

// Ingredient represents a recipe ingredient.
type Ingredient struct {
	Name   string  `bson:"name"   json:"name"`
	Amount float64 `bson:"amount" json:"amount"`
	Unit   string  `bson:"unit"   json:"unit"`
}

// Step represents a cooking step.
type Step struct {
	Order       int    `bson:"order"       json:"order"`
	Description string `bson:"description" json:"description"`
	Duration    int    `bson:"duration"    json:"duration"` // in minutes
}

// Validate performs business validation on the recipe.
func (r *Recipe) Validate() error {
	if r.Title == "" {
		return ErrEmptyTitle
	}
	if len(r.Title) < 3 || len(r.Title) > 200 {
		return ErrInvalidTitleLength
	}
	if r.Description == "" {
		return ErrEmptyDescription
	}
	if len(r.Description) < 10 || len(r.Description) > 5000 {
		return ErrInvalidDescriptionLength
	}
	if len(r.Ingredients) == 0 {
		return ErrNoIngredients
	}
	if len(r.Steps) == 0 {
		return ErrNoSteps
	}
	if r.CookingTime <= 0 {
		return ErrInvalidCookingTime
	}
	if r.Portions < 1 || r.Portions > 100 {
		return ErrInvalidPortions
	}
	if len(r.Tags) == 0 {
		return ErrNoTags
	}
	return nil
}

// SetTimestamps sets created and updated timestamps.
func (r *Recipe) SetTimestamps() {
	now := time.Now()
	if r.CreatedAt.IsZero() {
		r.CreatedAt = now
	}
	r.UpdatedAt = now
}
