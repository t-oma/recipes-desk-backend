package handler

import (
	"time"

	"recipes-desk/internal/modules/auth/service"
)

func toUserResponse(user *service.SafeUser) UserResponse {
	return UserResponse{
		ID:                user.ID.Hex(),
		Email:             user.Email,
		FirstName:         user.FirstName,
		LastName:          user.LastName,
		CreatedAt:         user.CreatedAt.Format(time.RFC3339),
		PasswordUpdatedAt: user.PasswordUpdatedAt.Format(time.RFC3339),
	}
}

func toResponse(authResult *service.AuthResult) AuthResponse {
	return AuthResponse{
		User: toUserResponse(authResult.User),
	}
}
