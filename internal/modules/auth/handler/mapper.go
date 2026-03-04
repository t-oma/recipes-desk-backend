package handler

import (
	"time"

	"recipes-desk/internal/modules/auth/application/dto"
)

func toUserResponse(user *dto.User) UserResponse {
	return UserResponse{
		ID:                user.ID,
		Email:             user.Email,
		FirstName:         user.FirstName,
		LastName:          user.LastName,
		CreatedAt:         user.CreatedAt.Format(time.RFC3339),
		PasswordUpdatedAt: user.PasswordUpdatedAt.Format(time.RFC3339),
	}
}

func toAuthResponse(authResult *dto.AuthResult) AuthResponse {
	return AuthResponse{
		User: toUserResponse(authResult.User),
	}
}
