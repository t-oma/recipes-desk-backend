package mapper

import (
	"recipes-desk/internal/modules/tags/application/dto"
	"recipes-desk/internal/modules/tags/domain/entity"
)

func ToTagDTO(tag *entity.Tag) *dto.Tag {
	return &dto.Tag{
		ID:        tag.ID().String(),
		Name:      tag.Name().String(),
		Slug:      tag.Slug().String(),
		CreatedAt: tag.CreatedAt(),
		UpdatedAt: tag.UpdatedAt(),
	}
}
