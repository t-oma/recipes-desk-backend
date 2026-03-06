package mapper

import (
	"recipes-desk/internal/modules/auth/application/dto"
	"recipes-desk/internal/modules/auth/domain/entity"
)

func ToUserDTO(user *entity.User) *dto.User {
	return &dto.User{
		ID:                user.ID().String(),
		Email:             user.Email().String(),
		FirstName:         user.FirstName().String(),
		LastName:          user.LastName().String(),
		CreatedAt:         user.CreatedAt(),
		PasswordUpdatedAt: user.PasswordUpdatedAt(),
	}
}
