package mongorepo

import "recipes-desk/internal/modules/recipes/domain/valueobject"

type stepModel struct {
	Order       int    `bson:"order"`
	Description string `bson:"description"`
	Duration    int64  `bson:"duration"` // in minutes
}

func (m *stepModel) toDomain() valueobject.Step {
	step, _ := valueobject.NewStep(m.Order, m.Description, m.Duration)
	return step
}

func stepModelFromDomain(step valueobject.Step) stepModel {
	return stepModel{
		Order:       step.Order(),
		Description: step.Description(),
		Duration:    step.SecondsInt64(),
	}
}
