package entity

import (
	"time"

	vo "recipes-desk/internal/modules/auth/domain/valueobject"
)

type User struct { //nolint:recvcheck // intentionally mixed pointer and value receivers
	id                vo.UserID
	email             vo.Email
	firstName         vo.FirstName
	lastName          vo.LastName
	passwordHash      vo.PasswordHash
	createdAt         time.Time
	passwordUpdatedAt time.Time
}

func NewUser(
	id vo.UserID,
	email vo.Email,
	firstName vo.FirstName,
	lastName vo.LastName,
	password vo.PasswordHash,
) *User {
	return &User{
		id:                id,
		email:             email,
		firstName:         firstName,
		lastName:          lastName,
		passwordHash:      password,
		createdAt:         time.Now(),
		passwordUpdatedAt: time.Now(),
	}
}

func (u *User) UpdateEmail(email vo.Email) {
	u.email = email
}

func (u *User) UpdateFirstName(firstName vo.FirstName) {
	u.firstName = firstName
}

func (u *User) UpdateLastName(lastName vo.LastName) {
	u.lastName = lastName
}

func (u *User) UpdatePassword(password vo.Password) {
	u.passwordHash = vo.PasswordHash(password.String())
}

func (u User) ID() vo.UserID {
	return u.id
}

func (u User) Email() vo.Email {
	return u.email
}

func (u User) FirstName() vo.FirstName {
	return u.firstName
}

func (u User) LastName() vo.LastName {
	return u.lastName
}

func (u User) PasswordHash() vo.PasswordHash {
	return u.passwordHash
}

func (u User) CreatedAt() time.Time {
	return u.createdAt
}

func (u User) PasswordUpdatedAt() time.Time {
	return u.passwordUpdatedAt
}

func (u User) HasID() bool {
	return u.id.String() != ""
}

func (u *User) AssignID(id vo.UserID) {
	u.id = id
}

func (u User) Equals(other *User) bool {
	if other == nil {
		return false
	}

	return u.id.String() == other.id.String()
}

func (u *User) RestoreFromPersistence(createdAt time.Time, passwordUpdatedAt time.Time) {
	u.createdAt = createdAt
	u.passwordUpdatedAt = passwordUpdatedAt
}
