package mongorepo

import "recipes-desk/internal/modules/recipes/domain/recipe"

type tagModel struct {
	Name string `bson:"name"`
}

func (m *tagModel) toDomain() recipe.Tag {
	tag, _ := recipe.NewTag(m.Name)
	return tag
}

func tagModelFromDomain(tag recipe.Tag) tagModel {
	return tagModel{
		Name: tag.Name(),
	}
}
