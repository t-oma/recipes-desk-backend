package mongorepo

import (
	"recipes-desk/internal/modules/recipes/domain/recipe"
)

type stepModel struct {
	Order       int    `bson:"order"`
	Description string `bson:"description"`
	Duration    int64  `bson:"duration"` // in minutes
}

func (m *stepModel) toDomain() recipe.Step {
	step, _ := recipe.NewStep(m.Order, m.Description, m.Duration)
	return step
}

func stepModelFromDomain(step recipe.Step) stepModel {
	return stepModel{
		Order:       step.Order(),
		Description: step.Description(),
		Duration:    step.SecondsInt64(),
	}
}
