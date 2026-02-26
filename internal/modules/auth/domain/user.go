package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID                primitive.ObjectID `bson:"_id"               json:"id"` //nolint:tagliatelle // mongoDB id
	Email             string             `bson:"email"             json:"email"`
	FirstName         string             `bson:"firstName"         json:"firstName"`
	LastName          string             `bson:"lastName"          json:"lastName"`
	Password          string             `bson:"password"          json:"password"`
	CreatedAt         time.Time          `bson:"createdAt"         json:"createdAt"`
	PasswordUpdatedAt time.Time          `bson:"passwordUpdatedAt" json:"passwordUpdatedAt"`
}

func (u *User) SetTimestamps() {
	now := time.Now()
	if u.CreatedAt.IsZero() {
		u.CreatedAt = now
	}
	if u.PasswordUpdatedAt.IsZero() {
		u.PasswordUpdatedAt = now
	}
}

// Validate performs business validation on the user.
func (u *User) Validate() error {
	if u.Email == "" {
		return ErrEmptyEmail
	}
	if len(u.Email) < 3 || len(u.Email) > 254 {
		return ErrInvalidEmailLength
	}
	if u.Password == "" {
		return ErrEmptyPassword
	}
	if len(u.Password) < 8 {
		return ErrPasswordTooShort
	}
	if len(u.Password) > 72 {
		return ErrPasswordTooLong
	}
	if u.FirstName == "" {
		return ErrEmptyFirstName
	}
	if len(u.FirstName) < 2 || len(u.FirstName) > 50 {
		return ErrInvalidFirstNameLength
	}
	if u.LastName == "" {
		return ErrEmptyLastName
	}
	if len(u.LastName) < 2 || len(u.LastName) > 50 {
		return ErrInvalidLastNameLength
	}
	return nil
}
