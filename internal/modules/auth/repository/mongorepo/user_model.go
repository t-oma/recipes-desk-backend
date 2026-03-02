package mongorepo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"recipes-desk/internal/modules/auth/domain"
)

// userModel represents a user in MongoDB.
type userModel struct {
	ID                primitive.ObjectID `bson:"_id,omitempty"` //nolint:tagliatelle // MongoDB _id field
	Email             string             `bson:"email"`
	FirstName         string             `bson:"firstName"`
	LastName          string             `bson:"lastName"`
	Password          string             `bson:"password"`
	CreatedAt         time.Time          `bson:"createdAt"`
	PasswordUpdatedAt time.Time          `bson:"passwordUpdatedAt"`
}

func (m *userModel) setTimestamps() {
	now := time.Now()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	if m.PasswordUpdatedAt.IsZero() {
		m.PasswordUpdatedAt = now
	}
}

func (m *userModel) setID() {
	m.ID = primitive.NewObjectID()
}

func (m *userModel) prepareForInsert() {
	m.setTimestamps()
	m.setID()
}

// toDomain converts a MongoDB user model to a domain user.
func (m *userModel) toDomain() *domain.User {
	return &domain.User{
		ID:                m.ID.Hex(),
		Email:             m.Email,
		FirstName:         m.FirstName,
		LastName:          m.LastName,
		Password:          m.Password,
		CreatedAt:         m.CreatedAt,
		PasswordUpdatedAt: m.PasswordUpdatedAt,
	}
}

// UserModelFromDomain converts a domain user to a MongoDB user model.
func userModelFromDomain(user *domain.User) *userModel {
	var id primitive.ObjectID
	if user.ID != "" {
		var err error
		id, err = primitive.ObjectIDFromHex(user.ID)
		if err != nil {
			id = primitive.NilObjectID
		}
	}

	return &userModel{
		ID:                id,
		Email:             user.Email,
		FirstName:         user.FirstName,
		LastName:          user.LastName,
		Password:          user.Password,
		CreatedAt:         user.CreatedAt,
		PasswordUpdatedAt: user.PasswordUpdatedAt,
	}
}
