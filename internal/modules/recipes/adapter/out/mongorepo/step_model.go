package mongorepo

import "recipes-desk/internal/modules/recipes/domain"

type stepModel struct {
	Order       int    `bson:"order"`
	Description string `bson:"description"`
	Duration    int    `bson:"duration"` // in minutes
}

func (m *stepModel) toDomain() domain.Step {
	return domain.Step{
		Order:       m.Order,
		Description: m.Description,
		Duration:    m.Duration,
	}
}

func stepModelFromDomain(step domain.Step) stepModel {
	return stepModel{
		Order:       step.Order,
		Description: step.Description,
		Duration:    step.Duration,
	}
}
