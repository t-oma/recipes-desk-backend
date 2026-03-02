package domain

import (
	"time"
)

// User represents a user in the domain layer.
type User struct {
	ID                string
	Email             string
	FirstName         string
	LastName          string
	Password          string
	CreatedAt         time.Time
	PasswordUpdatedAt time.Time
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
