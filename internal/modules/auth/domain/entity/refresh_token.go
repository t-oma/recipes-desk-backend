package entity

import (
	"time"

	vo "recipes-desk/internal/modules/auth/domain/valueobject"
)

type RefreshToken struct {
	id        vo.RefreshTokenID
	userID    vo.UserID
	tokenHash vo.TokenHash
	expiresAt time.Time
	createdAt time.Time
}

func NewRefreshToken(
	id vo.RefreshTokenID,
	userID vo.UserID,
	tokenHash vo.TokenHash,
) *RefreshToken {
	return &RefreshToken{
		id:        id,
		userID:    userID,
		tokenHash: tokenHash,
		expiresAt: time.Now(),
		createdAt: time.Now(),
	}
}

func (r *RefreshToken) SetID(id vo.RefreshTokenID) {
	r.id = id
}

func (r *RefreshToken) SetUserID(userID vo.UserID) {
	r.userID = userID
}

func (r *RefreshToken) SetTokenHash(tokenHash vo.TokenHash) {
	r.tokenHash = tokenHash
}

func (r *RefreshToken) ID() vo.RefreshTokenID {
	return r.id
}

func (r *RefreshToken) UserID() vo.UserID {
	return r.userID
}

func (r *RefreshToken) TokenHash() vo.TokenHash {
	return r.tokenHash
}

func (r *RefreshToken) ExpiresAt() time.Time {
	return r.expiresAt
}

func (r *RefreshToken) CreatedAt() time.Time {
	return r.createdAt
}

func (r *RefreshToken) RestoreFromPersistence(createdAt time.Time, expiresAt time.Time) {
	r.createdAt = createdAt
	r.expiresAt = expiresAt
}
