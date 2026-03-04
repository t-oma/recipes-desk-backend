package mapper

import (
	"recipes-desk/internal/modules/auth/application/dto"
	"recipes-desk/internal/modules/auth/domain"
)

func ToUserDTO(user *domain.User) *dto.User {
	return &dto.User{
		ID:                user.ID,
		Email:             user.Email,
		FirstName:         user.FirstName,
		LastName:          user.LastName,
		CreatedAt:         user.CreatedAt,
		PasswordUpdatedAt: user.PasswordUpdatedAt,
	}
}
