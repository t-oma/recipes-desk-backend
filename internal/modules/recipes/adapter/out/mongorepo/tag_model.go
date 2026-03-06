package mongorepo

import "recipes-desk/internal/modules/recipes/domain/valueobject"

type tagModel struct {
	Name string `bson:"name"`
}

func (m *tagModel) toDomain() valueobject.Tag {
	tag, _ := valueobject.NewTag(m.Name)
	return tag
}

func tagModelFromDomain(tag valueobject.Tag) tagModel {
	return tagModel{
		Name: tag.Name(),
	}
}
