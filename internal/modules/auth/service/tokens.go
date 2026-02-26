package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"recipes-desk/internal/modules/auth/domain"
)

var ErrSignToken = errors.New("failed to sign token")

type Claims struct {
	jwt.RegisteredClaims

	UserID string `json:"userId"`
}

// TokenService handles JWT token generation and validation.
type TokenService struct {
	refreshRepo domain.RefreshTokenRepository
	secret      []byte
	accessTTL   time.Duration
	refreshTTL  time.Duration
}

func NewTokenService(
	refreshRepo domain.RefreshTokenRepository,
	secret string,
	accessTTL time.Duration,
	refreshTTL time.Duration,
) *TokenService {
	return &TokenService{
		refreshRepo: refreshRepo,
		secret:      []byte(secret),
		accessTTL:   accessTTL,
		refreshTTL:  refreshTTL,
	}
}

// GenerateAccessToken creates a new JWT access token for the user.
func (s *TokenService) GenerateAccessToken(userID string) (*TokenResult, error) {
	tokenString, expiresAt, err := GenerateToken(
		s.secret,
		jwt.SigningMethodHS256,
		userID,
		s.accessTTL,
	)
	if err != nil {
		return nil, err
	}

	return &TokenResult{
		Token:     tokenString,
		ExpiresAt: expiresAt,
	}, nil
}

// ValidateAccessToken validates the JWT access token and returns the claims.
func (s *TokenService) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{}, //nolint:exhaustruct // JWT library initializes fields
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return s.secret, nil
		})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, domain.ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidToken, err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, domain.ErrInvalidToken
}

// GenerateRefreshToken creates a new refresh token and stores its hash in the database.
func (s *TokenService) GenerateRefreshToken(
	ctx context.Context,
	userID primitive.ObjectID,
) (*TokenResult, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random token: %w", err)
	}

	plainToken := base64.URLEncoding.EncodeToString(randomBytes)
	tokenHash := hashToken(plainToken)

	refreshToken := &domain.RefreshToken{ //nolint:exhaustruct // fields set via SetTimestamps
		UserID:    userID,
		TokenHash: tokenHash,
	}
	refreshToken.SetTimestamps(s.refreshTTL)

	if err := s.refreshRepo.Create(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &TokenResult{
		Token:     plainToken,
		ExpiresAt: refreshToken.ExpiresAt,
	}, nil
}

// ValidateRefreshToken validates a refresh token and returns the associated user ID.
func (s *TokenService) ValidateRefreshToken(
	ctx context.Context,
	plainToken string,
) (primitive.ObjectID, error) {
	tokenHash := hashToken(plainToken)

	storedToken, err := s.refreshRepo.FindByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return primitive.NilObjectID, domain.ErrTokenNotFound
		}
		return primitive.NilObjectID, fmt.Errorf("failed to find refresh token: %w", err)
	}

	// Check if token has expired
	if time.Now().After(storedToken.ExpiresAt) {
		return primitive.NilObjectID, domain.ErrExpiredToken
	}

	return storedToken.UserID, nil
}

// RotateRefreshToken invalidates the old refresh token and creates a new one.
func (s *TokenService) RotateRefreshToken(
	ctx context.Context,
	oldPlainToken string,
) (*TokenResult, primitive.ObjectID, error) {
	userID, err := s.ValidateRefreshToken(ctx, oldPlainToken)
	if err != nil {
		return nil, primitive.NilObjectID, err
	}

	oldTokenHash := hashToken(oldPlainToken)
	if err = s.refreshRepo.DeleteByHash(ctx, oldTokenHash); err != nil {
		return nil, primitive.NilObjectID, fmt.Errorf(
			"failed to delete old refresh token: %w",
			err,
		)
	}

	// Generate new token
	newToken, err := s.GenerateRefreshToken(ctx, userID)
	if err != nil {
		return nil, primitive.NilObjectID, err
	}

	return newToken, userID, nil
}

// hashToken creates a SHA-256 hash of the token.
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.StdEncoding.EncodeToString(hash[:])
}

func GenerateToken(
	secret any,
	method jwt.SigningMethod,
	userID string,
	ttl time.Duration,
) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(ttl)
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{ //nolint:exhaustruct // only need ExpiresAt, IssuedAt, NotBefore
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(method, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("%w: %w", ErrSignToken, err)
	}

	return tokenString, expiresAt, nil
}
