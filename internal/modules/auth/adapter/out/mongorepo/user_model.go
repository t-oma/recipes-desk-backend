package mongorepo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"recipes-desk/internal/modules/auth/domain/entity"
	"recipes-desk/internal/modules/auth/domain/valueobject"
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
func (m *userModel) toDomain() (*entity.User, error) {
	idVO, err := valueobject.NewUserID(m.ID.Hex())
	if err != nil {
		return nil, err
	}
	emailVO, err := valueobject.NewEmail(m.Email)
	if err != nil {
		return nil, err
	}
	firstNameVO, err := valueobject.NewFirstName(m.FirstName)
	if err != nil {
		return nil, err
	}
	lastNameVO, err := valueobject.NewLastName(m.LastName)
	if err != nil {
		return nil, err
	}

	user := entity.NewUser(
		idVO,
		emailVO,
		firstNameVO,
		lastNameVO,
		valueobject.PasswordHash(m.Password),
	)
	user.RestoreFromPersistence(m.CreatedAt, m.PasswordUpdatedAt)
	return user, nil
}

// UserModelFromDomain converts a domain user to a MongoDB user model.
func userModelFromDomain(user *entity.User) *userModel {
	var id primitive.ObjectID
	if user.HasID() {
		var err error
		id, err = primitive.ObjectIDFromHex(user.ID().String())
		if err != nil {
			id = primitive.NilObjectID
		}
	}

	return &userModel{
		ID:                id,
		Email:             user.Email().String(),
		FirstName:         user.FirstName().String(),
		LastName:          user.LastName().String(),
		Password:          string(user.PasswordHash()),
		CreatedAt:         user.CreatedAt(),
		PasswordUpdatedAt: user.PasswordUpdatedAt(),
	}
}
