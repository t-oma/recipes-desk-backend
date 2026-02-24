package handler

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"recipes-desk/internal/modules/auth/domain"
	"recipes-desk/internal/modules/auth/service"
)

func toDomainUser(req RegisterRequest) *domain.User {
	return &domain.User{
		ID:                primitive.NilObjectID,
		Email:             req.Email,
		FirstName:         req.FirstName,
		LastName:          req.LastName,
		Password:          req.Password,
		CreatedAt:         time.Time{},
		PasswordUpdatedAt: time.Time{},
	}
}

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
