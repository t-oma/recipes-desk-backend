package mongorepo

import "recipes-desk/internal/modules/recipes/domain/valueobject"

type tagModel struct {
	Name string `bson:"name"`
}

func (m *tagModel) toDomain() (valueobject.Tag, error) {
	tag, err := valueobject.NewTag(m.Name)
	return tag, err
}

func tagModelFromDomain(tag valueobject.Tag) tagModel {
	return tagModel{
		Name: tag.Name(),
	}
}
