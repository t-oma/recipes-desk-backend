package mongorepo

import "recipes-desk/internal/modules/recipes/domain/valueobject"

type tagNameModel struct {
	Name string `bson:"name"`
}

func (m *tagNameModel) toDomain() (valueobject.TagName, error) {
	tag, err := valueobject.NewTagName(m.Name)
	return tag, err
}

func tagModelFromDomain(tag valueobject.TagName) tagNameModel {
	return tagNameModel{
		Name: tag.String(),
	}
}
